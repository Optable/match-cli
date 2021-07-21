package auth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"
)

type EphemerealCertificate struct {
	CertificatePem []byte
	PrivateKeyPem  []byte
}

func NewEphemerealCertificate(privateKey *ecdsa.PrivateKey) (*EphemerealCertificate, error) {
	ret := EphemerealCertificate{}

	privateKeyDer, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	ret.PrivateKeyPem = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyDer,
	})

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemereal certificate serial number: %w", err)
	}
	ephemerealTemplate := &x509.Certificate{
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		SerialNumber:          serialNumber,
		NotBefore:             time.Now().Add(-30 * time.Minute),
		NotAfter:              time.Now().Add(60 * time.Minute),
		BasicConstraintsValid: true,
	}
	certificateDer, err := x509.CreateCertificate(rand.Reader, ephemerealTemplate, ephemerealTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create x509 certificate: %w", err)
	}

	ret.CertificatePem = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificateDer,
	})
	return &ret, nil
}

func (c *EphemerealCertificate) GetTLSCertificate() (tls.Certificate, error) {
	return tls.X509KeyPair(c.CertificatePem, c.PrivateKeyPem)
}

func ParseCertificatePEM(certificatePEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 certificate: %w", err)
	}
	return cert, nil
}

type PeerCertificateVerifier func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

// MakeVerifyPinnedCertificate verifies the peer certificates on the TLS handshake for one that
// stricly matches a previously shared pinned certificate.
// We use it to verify ephemereal certificates exchanged through a side channel.
func MakeVerifyPinnedCertificate(pinnedCert *x509.Certificate) PeerCertificateVerifier {
	return func(rawCerts [][]byte, _verifiedChains [][]*x509.Certificate) error {
		for _, c := range rawCerts {
			peerCert, err := x509.ParseCertificate(c)
			if err != nil {
				return fmt.Errorf("failed to parse peer certificate: %w", err)
			}

			if pinnedCert.Equal(peerCert) {
				return nil
			}
		}

		return errors.New("failed to find a peer certificate that matches the pinned certificate")
	}
}
