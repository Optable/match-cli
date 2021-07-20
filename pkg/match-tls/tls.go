package match_tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

const network = "tcp"

// Connect establish a tls connection to the endpoint with nagle enabled.
func Connect(ctx context.Context, endpoint string, cred *tls.Config) (*tls.Conn, error) {
	raddr, err := net.ResolveTCPAddr(network, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed resolving TCP address of %s: %w", endpoint, err)
	}

	// maximum time for retrying tls connection.
	timeout := time.NewTimer(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			return nil, fmt.Errorf("TLS connection time exceeded")
		default:
			// retry connection
		}

		conn, err := net.DialTCP(network, nil, raddr)
		if err != nil {
			return nil, fmt.Errorf("failed to dial %s: %w", endpoint, err)
		}
		// enable nagle
		if err := conn.SetNoDelay(false); err != nil {
			return nil, fmt.Errorf("cannot enable nagle: %w", err)
		}

		c := tls.Client(conn, cred)
		if err := c.Handshake(); err != nil {
			// wait 1 second and retry
			time.Sleep(time.Second)
			continue
		}

		return c, nil
	}
}

// Listen listens to a tcp connection to host and returns a nagle enabled tls connection
func Listen(host string, cred *tls.Config) (*tls.Conn, error) {
	laddr, err := net.ResolveTCPAddr(network, host)
	if err != nil {
		return nil, fmt.Errorf("failed resolving TCP address of %s: %w", host, err)
	}

	l, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen %w", err)
	}

	conn, err := l.AcceptTCP()
	if err != nil {
		return nil, fmt.Errorf("failed to accept connection: %w", err)
	}
	// enable nagle
	if err := conn.SetNoDelay(false); err != nil {
		return nil, fmt.Errorf("cannot enable nagle: %w", err)
	}

	c := tls.Server(conn, cred)
	if err := c.Handshake(); err != nil {
		return nil, fmt.Errorf("server: TLS handshake failed: %w", err)
	}

	return c, nil
}
