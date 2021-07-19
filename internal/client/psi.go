package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	match_header "github.com/optable/match-cli/pkg/match-header"
	"github.com/optable/match/pkg/psi"

	"github.com/rs/zerolog"
)

func Send(ctx context.Context, endpoint string, creds *tls.Config, n int64, in <-chan []byte) error {
	c, err := connect(ctx, endpoint, creds)
	if err != nil {
		return err
	}
	zerolog.Ctx(ctx).Info().Msgf("connected to partner")

	// protocol negotiation step
	protocol, err := match_header.NegotiateSenderProtocol(c)
	if err != nil {
		return err
	}
	zerolog.Ctx(ctx).Info().Msgf("received protocol: %d", protocol)

	sender, err := psi.NewSender(protocol, c)
	if err != nil {
		return fmt.Errorf("failed creating PSI sender %w", err)
	}

	zerolog.Ctx(ctx).Info().Msgf("created sender to start PSI")

	return sender.Send(ctx, n, in)
}

// connect establish a tls connection to the endpoint with nagle enabled.
func connect(ctx context.Context, endpoint string, cred *tls.Config) (*tls.Conn, error) {
	raddr, err := net.ResolveTCPAddr("tcp", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed resolving TCP address of %s: %w", endpoint, err)
	}

	timeout := time.NewTimer(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			return nil, fmt.Errorf("connection time exceeded")
		default:
			// try connection
		}

		conn, err := net.DialTCP("tcp", nil, raddr)
		if err != nil {
			return nil, fmt.Errorf("failed to dial %s: %w", endpoint, err)
		}
		// enable nagle
		if err := conn.SetNoDelay(false); err != nil {
			return nil, fmt.Errorf("cannot enable nagle: %w", err)
		}

		c := tls.Client(conn, cred)
		if err := c.Handshake(); err != nil {
			time.Sleep(time.Second)
			continue
		}

		return c, nil
	}
}
