package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	match "optable-sandbox/pkg/lib/match-header"

	"github.com/optable/match/pkg/psi"
	"github.com/rs/zerolog"
)

func RunPSI(ctx context.Context, endpoint string, creds *tls.Config, n int64, in <-chan []byte) error {
	c, err := connect(ctx, endpoint, creds)
	if err != nil {
		return err
	}
	zerolog.Ctx(ctx).Info().Msgf("connected to partner")

	// protocol negotiation step
	protocol, err := match.NegotiateSenderProtocol(c)
	if err != nil {
		return err
	}
	zerolog.Ctx(ctx).Info().Msgf("received protocol: %d", protocol)

	sender, err := psi.NewSender(protocol, c)
	if err != nil {
		return fmt.Errorf("Failed creating PSI sender %w", err)
	}

	zerolog.Ctx(ctx).Info().Msgf("created sender to start PSI")

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

// negotiateSenderProtocol receives server supported protocols and selects one.
func negotiateSenderProtocol(rw io.ReadWriter) (int, error) {
	protocol := make([]byte, 1)
	if n, err := rw.Read(protocol); err != nil || n != len(protocol) {
		return n, fmt.Errorf("Sender failed to receive PSI protocol negotiation message, got: %v, err: %w", protocol, err)
	}

	return int(protocol[0]), nil
}
