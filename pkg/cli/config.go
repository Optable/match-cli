package cli

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/optable/match-cli/internal/client"

	"github.com/golang-jwt/jwt/v4"
)

const configFile = ".config/optable/optable-match-cli.conf"

func getConfigPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir + "/" + configFile, nil
}

func ensureConfigPath() (string, error) {
	path, err := getConfigPath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return path, nil
}

type Config struct {
	Partners []PartnerConfig `json:"partners"`
}

func (c *Config) findPartner(name string) *PartnerConfig {
	for _, p := range c.Partners {
		if p.Name == name {
			return &p
		}
	}

	return nil
}

type PartnerConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Id          string `json:"id"`
	PrivateKey  string `json:"private_key"`
}

func (partner *PartnerConfig) ParsedPrivateKey() (*ecdsa.PrivateKey, error) {
	privateKeyDer, err := base64.StdEncoding.DecodeString(partner.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 encoded private key: %w", err)
	}

	parsedPrivateKey, err := x509.ParseECPrivateKey(privateKeyDer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}
	return parsedPrivateKey, nil
}

func (partner *PartnerConfig) NewToken(expireAt time.Duration) (string, error) {
	claims := jwt.StandardClaims{
		Issuer:    partner.Id,
		ExpiresAt: time.Now().Add(expireAt).Unix(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	parsedKey, err := partner.ParsedPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	tokStr, err := tok.SignedString(parsedKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokStr, nil
}

func (partner *PartnerConfig) NewClient() (*client.AdminRpcClient, error) {
	tokenSourceFn := func(_ *http.Request) (string, error) {
		return partner.NewToken(time.Minute * 10)
	}
	return client.NewClient(partner.URL, client.TokenSourceFn(tokenSourceFn)), nil
}
