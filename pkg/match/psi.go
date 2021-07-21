package match

import (
	"context"
	"crypto/tls"
	"fmt"

	header "github.com/optable/match-cli/pkg/header"
	network "github.com/optable/match-cli/pkg/network"
	"github.com/optable/match/pkg/psi"

	"github.com/rs/zerolog"
)

// Send initiate a tls connection with the match receiver,
// negotiate and establish a PSI protocol
// instantiate and act as a sender in the specified PSI protocol,
// and returns any error encountered during the match.
func Send(ctx context.Context, endpoint string, creds *tls.Config, n int64, in <-chan []byte) error {
	c, err := network.Connect(ctx, endpoint, creds)
	if err != nil {
		return err
	}
	zerolog.Ctx(ctx).Info().Msgf("connected to partner")

	// protocol negotiation step
	protocol, err := header.NegotiateSenderProtocol(c)
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

// Receive initiate a tls connection with the match sender,
// negotiate and establish a PSI protocol,
// instantiate and act as receiver in the specified PSI protocol,
// and returns the intersected identifiers or any errors encountered
// during the match.
func Receive(ctx context.Context, host string, cred *tls.Config, inputLen int64, identifiers <-chan []byte, protocols []uint8) ([][]byte, error) {
	c, err := network.Listen(host, cred)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// protocol negatiation step
	selected, err := header.NegotiateReceiverProtocol(c, protocols)
	if err != nil {
		return nil, err
	}

	receiver, err := psi.NewReceiver(selected, c)
	if err != nil {
		return nil, fmt.Errorf("failed creating PSI receiver %w", err)
	}

	// intersect
	return receiver.Intersect(ctx, inputLen, identifiers)
}
