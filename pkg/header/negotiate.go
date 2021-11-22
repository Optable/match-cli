package header

import (
	"fmt"
	"io"

	"github.com/optable/match/pkg/psi"
)

// NegotiateSenderProtocol sends the desired protocol to the receiver.
// The receiver responds with the same protocol if it is accepted.
// If the protocol is not accepted, the receiver will send back the default protocol.
func NegotiateSenderProtocol(rw io.ReadWriter, protocol psi.Protocol) (psi.Protocol, error) {
	if n, err := rw.Write([]uint8{uint8(protocol)}); err != nil || n != 1 {
		return protocol, fmt.Errorf("failed to send protocol negotiation message: %w", err)
	}
	recProtocol := make([]byte, 1)
	if _, err := rw.Read(recProtocol); err != nil {
		return protocol, fmt.Errorf("failed to receive PSI protocol negotiation message, got: %v, err: %w", protocol, err)
	}

	return psi.Protocol(recProtocol[0]), nil
}
