package header

import (
	"fmt"
	"io"

	"github.com/optable/match/pkg/psi"
)

// NegotiateSenderProtocol sends the desired protocol to the receiver.
// The receiver responds whether the protocol is accepted or not.
// If the protocol is not accepted, fall back to DHPSI.
func NegotiateSenderProtocol(rw io.ReadWriter, protocol psi.Protocol) (psi.Protocol, error) {
	if n, err := rw.Write([]uint8{uint8(protocol)}); err != nil || n != 1 {
		return protocol, fmt.Errorf("failed to send protocol negotiation message: %w", err)
	}
	ack := make([]byte, 1)
	if _, err := rw.Read(ack); err != nil {
		return protocol, fmt.Errorf("sender failed to receive PSI protocol negotiation message, got: %v, err: %w", protocol, err)
	}

	// Receiver doesn't accept proposed PSI so fall back on DHPSI
	if ack[0] == 0 {
		return psi.ProtocolDHPSI, nil
	}

	return protocol, nil
}
