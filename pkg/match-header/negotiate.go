package match_header

import (
	"fmt"
	"io"
)

// NegotiateSenderProtocol receives server supported protocols and selects one.
func NegotiateSenderProtocol(rw io.ReadWriter) (int, error) {
	protocol := make([]byte, 1)
	if n, err := rw.Read(protocol); err != nil || n != len(protocol) {
		return n, fmt.Errorf("sender failed to receive PSI protocol negotiation message, got: %v, err: %w", protocol, err)
	}

	return int(protocol[0]), nil
}

// NegotiateReceiverProtocol sends server supported protocols in order of preference and returns the agreed one.
func NegotiateReceiverProtocol(rw io.ReadWriter, protocols []uint8) (int, error) {
	protocolRes := []byte{protocols[0]}
	if n, err := rw.Write(protocolRes); err != nil || n != len(protocolRes) {
		return n, fmt.Errorf("failed to send protocol negotiation message: %w", err)
	}

	return int(protocols[0]), nil
}
