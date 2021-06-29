package crypt

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func ParseECPrivateKeyPEM(key string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("Failed to decode key pem")
	}
	parsedKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return parsedKey, nil
}
