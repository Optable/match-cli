package network

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"
)

const network = "tcp"

// Connect establish a tls connection to the endpoint with nagle enabled.
func Connect(ctx context.Context, endpoint string, cred *tls.Config) (*tls.Conn, error) {
	var conn *tls.Conn

	connectBlock := func() error {
		for {
			// dial timeout of 2 seconds.
			dialConn, err := net.DialTimeout(network, endpoint, 2*time.Second)
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				//retry
				continue
			}

			if err != nil {
				return fmt.Errorf("failed to dial %s: %w", endpoint, err)
			}

			// enable nagle
			tcpConn := dialConn.(*net.TCPConn)
			if err := tcpConn.SetNoDelay(false); err != nil {
				return fmt.Errorf("cannot enable nagle: %w", err)
			}

			tlsConn := tls.Client(tcpConn, cred)
			if err := tlsConn.Handshake(); err != nil {
				time.Sleep(time.Second)
				continue
			}

			conn = tlsConn
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if err := unblock(ctx, connectBlock); err != nil {
		return nil, err
	}

	return conn, nil
}

// Listen listens to a tcp connection to host and returns a nagle enabled tls connection
func Listen(ctx context.Context, host string, cred *tls.Config) (*tls.Conn, error) {
	laddr, err := net.ResolveTCPAddr(network, host)
	if err != nil {
		return nil, fmt.Errorf("failed resolving TCP address of %s: %w", host, err)
	}

	var conn *tls.Conn

	listenBlock := func() error {
		listen, err := net.ListenTCP(network, laddr)
		if err != nil {
			return fmt.Errorf("failed to listen %w", err)
		}

		tcpConn, err := listen.AcceptTCP()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}
		// enable nagle
		if err := tcpConn.SetNoDelay(false); err != nil {
			return fmt.Errorf("cannot enable nagle: %w", err)
		}

		tlsConn := tls.Server(tcpConn, cred)
		if err := tlsConn.Handshake(); err != nil {
			return fmt.Errorf("server: handshake failed: %w", err)
		}

		conn = tlsConn
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if err := unblock(ctx, listenBlock); err != nil {
		return nil, err
	}

	return conn, nil
}

func unblock(ctx context.Context, blocking func() error) error {
	if blocking == nil {
		return nil
	}

	err := make(chan error, 1)
	go func() {
		defer close(err)
		err <- blocking()
	}()

	select {
	case e := <-err:
		return e
	case <-ctx.Done():
		return ctx.Err()
	}
}
