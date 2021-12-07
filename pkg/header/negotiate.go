package header

import (
	"fmt"
	"io"

	"github.com/optable/match/pkg/psi"
)

// NegotiateSenderProtocol takes the sender's slice of protocols which
// are ordered in terms of desirability. First send the length of the
// protocols slice to receiver and then send the slice itself. The
// receiver should respond with the first element that is present in
// both the sender's and receiver's preferred protocol slices. If there
// is no intersection between the slices, the receiver will respond
// with the default protocol.
func NegotiateSenderProtocol(rw io.ReadWriter, protocols []psi.Protocol) (psi.Protocol, error) {
	// write length of preferred protocol slice
	if _, err := rw.Write([]byte{byte(len(protocols))}); err != nil {
		return psi.ProtocolUnsupported, fmt.Errorf("failed to send number of desired protocols: %w", err)
	}
	// write actual slice of preferred protocols
	protocolMessage := make([]byte, len(protocols))
	for p, i := range protocols {
		protocolMessage[i] = byte(p)
	}
	if _, err := rw.Write(protocolMessage); err != nil {
		return psi.ProtocolUnsupported, fmt.Errorf("failed to send protocol negotiation message: %w", err)
	}
	// read protocol decision from receiver
	protocolDecision := make([]byte, 1)
	if _, err := rw.Read(protocolDecision); err != nil {
		return psi.ProtocolUnsupported, fmt.Errorf("failed to receive PSI protocol decision, got: %v, err: %w", protocolDecision, err)
	}

	return psi.Protocol(protocolDecision[0]), nil
}
