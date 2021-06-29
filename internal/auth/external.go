package auth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/optable/match-cli/api/models"

	"github.com/dgrijalva/jwt-go"
)

type AuthenticationAttributes struct {
	Mechanism        AuthenticationMechanism
	AuthenticationID string
	RootCert         string
}

var errBadX5C = errors.New("Token header x5c must be an array of a single base64 encoded DER x509 certs")

type AuthenticationMechanism uint8

const (
	UnknownAuth AuthenticationMechanism = iota
	PublicKeyAuth
	HostnameAuth
)

func getTokenX5C(token *jwt.Token) (string, error) {
	arr, ok := token.Header["x5c"].([]interface{})
	if !ok || len(arr) != 1 {
		return "", errBadX5C
	}

	str, ok := arr[0].(string)
	if !ok {
		return "", errBadX5C
	}

	return str, nil
}

func parseTokenCertificateUnverified(token *jwt.Token) (*x509.Certificate, error) {
	str, err := getTokenX5C(token)
	if err != nil {
		return nil, err
	}

	der, _ := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, errBadX5C
	}

	return x509.ParseCertificate(der)
}

func parseTokenCertificate(token *jwt.Token, hostname string, rootCert string) (*x509.Certificate, error) {
	cert, err := parseTokenCertificateUnverified(token)
	if err != nil {
		return nil, err
	}

	if net.ParseIP(hostname) != nil {
		hostname = fmt.Sprintf("[%s]", hostname)
	}

	opts := x509.VerifyOptions{
		Roots:   x509.NewCertPool(),
		DNSName: hostname,
	}

	if ok := opts.Roots.AppendCertsFromPEM([]byte(rootCert)); !ok {
		return nil, errors.New("Failed to parse root certificate pem")
	}

	if _, err = cert.Verify(opts); err != nil {
		return nil, err
	}

	return cert, nil
}

func VerifyIssuer(token *jwt.Token, rootCert string, authID string, mechanism AuthenticationMechanism) (interface{}, error) {
	switch mechanism {
	case PublicKeyAuth:
		decodedPublicKey, err := base64.StdEncoding.DecodeString(authID)
		if err != nil {
			return nil, err
		}
		unmarshalledPublicKey, err := x509.ParsePKIXPublicKey(decodedPublicKey)
		if err != nil {
			return nil, err
		}
		return unmarshalledPublicKey, nil
	case HostnameAuth:
		cert, err := parseTokenCertificate(token, authID, rootCert)
		if err != nil {
			return nil, err
		}
		return cert.PublicKey, nil
	case UnknownAuth:
		fallthrough
	default:
		return nil, fmt.Errorf("Failed to authenticate the token, unknown mechanism: %d", mechanism)
	}
}

func ParseExternalToken(token string, verifier func(slug string) (AuthenticationAttributes, error)) (string, error) {
	claims := jwt.StandardClaims{}

	parser := jwt.Parser{
		ValidMethods: []string{jwt.SigningMethodES256.Alg()},
	}

	_, err := parser.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		attr, err := verifier(claims.Issuer)
		if err != nil {
			return nil, err
		}
		return VerifyIssuer(token, attr.RootCert, attr.AuthenticationID, attr.Mechanism)
	})

	if err != nil {
		return "", err
	}

	return claims.Issuer, nil
}

const MatchAuthTokenValidityDuration = 5 * time.Minute

func NewExternalToken(slug string, cert *pem.Block, key *ecdsa.PrivateKey) (string, error) {
	now := time.Now()
	expireAt := MatchAuthTokenValidityDuration

	return createExternalToken(slug, cert, key, now, expireAt)
}

func createExternalToken(slug string, cert *pem.Block, key *ecdsa.PrivateKey, now time.Time, expireAt time.Duration) (string, error) {
	base64Cert := base64.StdEncoding.EncodeToString(cert.Bytes)

	claims := jwt.StandardClaims{
		Issuer:    slug,
		ExpiresAt: now.Add(expireAt).Unix(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["x5c"] = []string{base64Cert}

	tokStr, err := tok.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokStr, nil
}

func CreateInitialExternalToken(sandboxInfo string, slug string, validFrom time.Time, validUntil time.Time) (*models.PartnersInitToken, error) {
	secretBytes := make([]byte, 256)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, err
	}
	secret := hex.EncodeToString(secretBytes)
	token := &models.PartnersInitToken{
		Slug:        slug,
		SandboxInfo: sandboxInfo,
		Secret:      secret,
		CreatedAt:   validFrom,
		ExpiresAt:   validUntil,
	}
	return token, nil
}
