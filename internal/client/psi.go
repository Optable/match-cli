package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/optable/match/pkg/psi"
)

func RunDHPSI(ctx context.Context, endpoint string, creds *tls.Config, n int64, in <-chan []byte) error {
	c, err := connect(ctx, endpoint, creds)

	sender, err := psi.NewSender(psi.NPSI, c)
	if err != nil {
		return fmt.Errorf("Failed creating PSI sender %w", err)
	}

	return sender.Send(ctx, n, in)
}

// connect establish a tls connection to the endpoint with nagle enabled.
func connect(ctx context.Context, endpoint string, cred *tls.Config) (*tls.Conn, error) {
	raddr, err := net.ResolveTCPAddr("tcp", endpoint)
	if err != nil {
		return nil, fmt.Errorf("Failed resolving TCP address of %s: %w", endpoint, err)
	}

	timeout := time.NewTimer(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			return nil, fmt.Errorf("Connection time exceeded.")
		default:
			// try connection
		}

		conn, err := net.DialTCP("tcp", nil, raddr)
		if err != nil {
			return nil, fmt.Errorf("Failed to dial %s: %w", endpoint, err)
		}
		// enable nagle
		if err := conn.SetNoDelay(false); err != nil {
			return nil, fmt.Errorf("Cannot enable nagle: %w", err)
		}

		c := tls.Client(conn, cred)
		if err := c.Handshake(); err != nil {
			time.Sleep(time.Second)
			continue
		}

		return c, nil
	}
}
