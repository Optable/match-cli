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
	// timeout for PSI pod to pod connection
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
				// if main context is timed out, abort operation
				return true, nil, fmt.Errorf("dial PSI timed out: %w", ctx.Err())
			default:
				//retry on any dial errors
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
			return true, nil, fmt.Errorf("client: failed to establish a secure tls connection: %w", err)
		}

		// succeed in establishing a tls connection
		return true, tlsConn, nil
	}

	for {
		done, tlsConn, err := tryConnect(ctx)
		if !done {
			// retry connection
			continue
		}

		return tlsConn, err
	}
}

// Listen listens on a host and returns a nagle enabled tls connection
func Listen(ctx context.Context, host string, cred *tls.Config) (*tls.Conn, error) {
	// timeout for PSI pod to pod connection
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, matchTCPNetwork, host)
	if err != nil {
		return nil, fmt.Errorf("failed to listen %w", err)
	}

	// unblock Accept when context deadline is reached.
	go func() {
		<-ctx.Done()
		// closing the listener will unblock Accept and make Accept return an error
		// since the context will always timeout
		// the listener will be closed as soon as either connectTimeout is reached
		// or the parent context's timeout is reached.
		listener.Close()
	}()

	conn, err := listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("failed to accept connection: %w", err)
	}

	// Disable TCP_NODELAY enables nagle's algorithm
	// which collects small packets and send them once
	// instead of sending each packet as soon as they are available.
	tcpConn := conn.(*net.TCPConn)
	if err := tcpConn.SetNoDelay(false); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable nagle: %w", err)
	}

	tlsConn := tls.Server(tcpConn, cred)
	// test the tls connection
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("server: failed to establish a secure tls connection: %w", err)
	}

	return tlsConn, nil
}
