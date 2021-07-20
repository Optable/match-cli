package header

import (
	"fmt"
	"io"
)

// NegotiateSenderProtocol receives a list of receiver supported PSI protocol and returns the selected one.
func NegotiateSenderProtocol(rw io.ReadWriter) (int, error) {
	protocol := make([]byte, 1)
	if n, err := rw.Read(protocol); err != nil || n != len(protocol) {
		return n, fmt.Errorf("sender failed to receive PSI protocol negotiation message, got: %v, err: %w", protocol, err)
	}

	return int(protocol[0]), nil
}

// NegotiateReceiverProtocol sends a list of supported PSI protocol with the sender and returns the selected one from the sender.
func NegotiateReceiverProtocol(rw io.ReadWriter, protocols []uint8) (int, error) {
	protocolRes := []byte{protocols[0]}
	if n, err := rw.Write(protocolRes); err != nil || n != len(protocolRes) {
		return n, fmt.Errorf("failed to send protocol negotiation message: %w", err)
	}

	return int(protocols[0]), nil
}
