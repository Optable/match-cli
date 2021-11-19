package match

import (
	"context"
	"crypto/tls"
	"fmt"

	header "github.com/optable/match-cli/pkg/header"
	network "github.com/optable/match-cli/pkg/network"
	"github.com/optable/match/pkg/psi"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
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
	log := zerolog.Ctx(ctx)
	log.Info().Msgf("connected to partner")

	// protocol negotiation step
	protocol, err := header.NegotiateSenderProtocol(c)
	if err != nil {
		return err
	}
	log.Info().Msgf("received protocol: %s", protocol)

	sender, err := psi.NewSender(protocol, c)
	if err != nil {
		return fmt.Errorf("failed creating PSI sender %w", err)
	}

	log.Info().Msgf("created sender to start PSI")

	// create zerologr and pass it to ctx
	logger := zerologr.New(log)

	// zerologr sets the global variable LevelFieldName to ""
	// see https://github.com/go-logr/zerologr/blob/master/zerologr.go#L76
	// this resets the change, and preserves the pretty printing.
	zerolog.LevelFieldName = "level"
	return sender.Send(logr.NewContext(ctx, logger), n, in)
}
