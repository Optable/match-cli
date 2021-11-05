package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

const (
	matchTCPNetwork = "tcp"
	connectTimeout  = 6 * time.Minute
	dialTimeout     = 2 * time.Second
)

// Connect establishes a tls connection to the endpoint with nagle enabled.
func Connect(ctx context.Context, endpoint string, cred *tls.Config) (*tls.Conn, error) {
	// We enforce a connection timeout that is relatively large to allow
	// time for any startup delay for the receiver after the initial PSI negotiation
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	// tryConnect returns a boolean to signal that we are done retrying connecting to
	// the listener, as well as the obtained TLS conn, and any errors that results from
	// connecting.
	tryConnect := func(ctx context.Context) (bool, *tls.Conn, error) {
		dialCtx, dialCancel := context.WithTimeout(ctx, dialTimeout)
		defer dialCancel()

		var dialer net.Dialer
		// DialContext returns an error if the context's timeout is reached.
		// this makes sure that Connect would not loop forever to retry
		// connections
		dialConn, err := dialer.DialContext(dialCtx, matchTCPNetwork, endpoint)
		if err != nil {
			select {
			case <-ctx.Done():
				// if main context has timed out, abort operation
				return true, nil, fmt.Errorf("dial PSI timed out: %w", ctx.Err())
			default:
				// retry on any dial errors
				return false, nil, nil
			}
		}

		// Disable TCP_NODELAY enables nagle's algorithm
		// which collects small packets and send them once
		// instead of sending each packet as soon as they are available.
		tcpConn := dialConn.(*net.TCPConn)
		if err = tcpConn.SetNoDelay(false); err != nil {
			dialConn.Close()
			return true, nil, fmt.Errorf("failed to enable nagle: %w", err)
		}

		tlsConn := tls.Client(tcpConn, cred)
		// test the tls connection
		if err = tlsConn.Handshake(); err != nil {
			dialConn.Close()
			// Retry on client tls handshake failures since they can
			// happen if the connection goes through a proxy that eagerly
			// reads from the stream (ex. SNI) and the receiver is not ready.
			return false, nil, nil
		}

		return true, tlsConn, nil
	}

	for {
		done, tlsConn, err := tryConnect(ctx)
		if !done {
			continue
		}

		return tlsConn, err
	}
}
