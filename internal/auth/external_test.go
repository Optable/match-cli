package auth

import (
	"crypto/ecdsa"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"github.com/optable/match-cli/pkg/crypt"

	"github.com/stretchr/testify/assert"
)

const issuer = "client.cloud.optable.co"

const rootPEM = `
-----BEGIN CERTIFICATE-----
MIIBCzCBsgIJAM+zVvpnM+VCMAoGCCqGSM49BAMCMA0xCzAJBgNVBAYTAkNBMCAX
DTIwMTEzMDA0MTIxNFoYDzIwNzUwOTAzMDQxMjE0WjANMQswCQYDVQQGEwJDQTBZ
MBMGByqGSM49AgEGCCqGSM49AwEHA0IABFh3Z0j90CZL2tIyB/nD/qYTUBfc+x+2
SZsSMmtcclv5Cr62r/H2pduodVqH6MdXroOeRmlA/WMLt3K59LYUlVgwCgYIKoZI
zj0EAwIDSAAwRQIgPuT8cyAX/Z8YYGnM/46WWNDUxD2CVbmJHQI1OPNIIMgCIQDU
MzhEyN+q2crlsSMBx2tpNqtGIPlxdwwU1+Ujo0N0Sw==
-----END CERTIFICATE-----`

const certPEM = `
-----BEGIN CERTIFICATE-----
MIIBNzCB36ADAgECAgkAkkWyuNDvIM0wCgYIKoZIzj0EAwIwDTELMAkGA1UEBhMC
Q0EwIBcNMjAxMTMwMDQxMjE0WhgPMjA3NTA5MDMwNDEyMTRaMA0xCzAJBgNVBAYT
AkNBMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESddX8frfPkMZx7wujXUt8lgp
jo1xcDMJqDWYGdaT/YLbPBTykgsozv748+elcPdM96asPskVxh3YaqCBA1ftzaMm
MCQwIgYDVR0RBBswGYIXY2xpZW50LmNsb3VkLm9wdGFibGUuY28wCgYIKoZIzj0E
AwIDRwAwRAIgMZbbmUOIWHMVPTgjXFYIjRsdSwEkHMzKq8WrYBRFCyQCICNWlais
lnh/oMprH2QMhhqIERXlQ37JN5iqTKy4eswy
-----END CERTIFICATE-----`

const privPEM = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIDnLYDlZaNwReGPHrk9Ysjx74hAJma65eztg0xHDjZ78oAoGCCqGSM49
AwEHoUQDQgAESddX8frfPkMZx7wujXUt8lgpjo1xcDMJqDWYGdaT/YLbPBTykgso
zv748+elcPdM96asPskVxh3YaqCBA1ftzQ==
-----END EC PRIVATE KEY-----`

const goodPublicKey = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESddX8frfPkMZx7wujXUt8lgpjo1xcDMJqDWYGdaT/YLbPBTykgsozv748+elcPdM96asPskVxh3YaqCBA1ftzQ=="

func buildToken(issuer string, block *pem.Block, key *ecdsa.PrivateKey) string {
	token, err := NewExternalToken(issuer, block, key)
	if err != nil {
		panic(err)
	}
	return token
}

// Helper function for parsing certificate
func getParsedCertificate(certPEM string) (*pem.Block, error) {
	parsedCert, _ := pem.Decode([]byte(certPEM))
	if parsedCert == nil {
		return nil, errors.New("Failed to decode certificate pem")
	}
	return parsedCert, nil
}

func createVerifyIssuer(authID string, mechanism AuthenticationMechanism) func(string) (AuthenticationAttributes, error) {
	f := func(slug string) (AuthenticationAttributes, error) {
		return AuthenticationAttributes{Mechanism: mechanism, AuthenticationID: authID, RootCert: rootPEM}, nil
	}
	return f
}

func TestValidateAccessRequestOk(t *testing.T) {
	parsedPrivPEM, _ := crypt.ParseECPrivateKeyPEM(privPEM)
	parsedCertPEM, _ := getParsedCertificate(certPEM)
	token := buildToken(issuer, parsedCertPEM, parsedPrivPEM)
	result, err := ParseExternalToken(token, createVerifyIssuer(issuer, HostnameAuth))
	assert.NoError(t, err)
	assert.Equal(t, issuer, result)
}

func TestExpiredButOtherwiseValidToken(t *testing.T) {
	parsedPrivPEM, _ := crypt.ParseECPrivateKeyPEM(privPEM)
	parsedCertPEM, _ := getParsedCertificate(certPEM)

	token, _ := createExternalToken(issuer, parsedCertPEM, parsedPrivPEM, time.Now().Add(-24*time.Hour), 1*time.Minute)
	_, err := ParseExternalToken(token, createVerifyIssuer(issuer, HostnameAuth))
	assert.EqualError(t, err, "token is expired by 23h59m0s")
}

const badHostIssuer = "myhost.mydomain.com"

func TestValidateAccessRequestBadHostIssuer(t *testing.T) {
	parsedPrivPEM, _ := crypt.ParseECPrivateKeyPEM(privPEM)
	parsedCertPEM, _ := getParsedCertificate(certPEM)

	token := buildToken(badHostIssuer, parsedCertPEM, parsedPrivPEM)
	_, err := ParseExternalToken(token, createVerifyIssuer(badHostIssuer, HostnameAuth))
	assert.EqualError(t, err, "x509: certificate is valid for client.cloud.optable.co, not myhost.mydomain.com")
}

func TestValidateAccessRequestBadPublicKeyIssuer(t *testing.T) {
	parsedPrivPEM, _ := crypt.ParseECPrivateKeyPEM(wrongPrivateKey)
	parsedCertPEM, _ := getParsedCertificate(certPEM)

	token := buildToken("someBadIssuer", parsedCertPEM, parsedPrivPEM)
	_, err := ParseExternalToken(token, createVerifyIssuer(goodPublicKey, PublicKeyAuth))
	assert.EqualError(t, err, "crypto/ecdsa: verification error")
}

func TestValidateAccessRequestGoodPublicKeyIssuer(t *testing.T) {
	parsedPrivPEM, _ := crypt.ParseECPrivateKeyPEM(privPEM)
	parsedCertPEM, _ := getParsedCertificate(certPEM)

	token := buildToken("someGoodIssuer", parsedCertPEM, parsedPrivPEM)
	result, err := ParseExternalToken(token, createVerifyIssuer(goodPublicKey, PublicKeyAuth))
	assert.NoError(t, err)
	assert.Equal(t, "someGoodIssuer", result)
}

const wrongPrivateKey = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIA4RDx/pkbnT0Ys/KzNkD4aWk7X9ImJZwWTY8MdKWI3koAoGCCqGSM49
AwEHoUQDQgAEPT7P89tEwfGbApP5uwb9Q8sTyidpmwXSAfo5amrEWeKvWxqeOXy1
M3YnKD2jr+fsFR9fkj/6SB0eMuu+fvRONw==
-----END EC PRIVATE KEY-----`

func TestValidateAccessRequestBadToken(t *testing.T) {
	parsedWrongPrivateKey, _ := crypt.ParseECPrivateKeyPEM(wrongPrivateKey)
	parsedCertPEM, _ := getParsedCertificate(certPEM)

	token := buildToken(issuer, parsedCertPEM, parsedWrongPrivateKey)
	_, err := ParseExternalToken(token, createVerifyIssuer(issuer, HostnameAuth))
	assert.EqualError(t, err, "crypto/ecdsa: verification error")
}

const certPEMWithIP = `
-----BEGIN CERTIFICATE-----
MIIBJTCBzKADAgECAgkAkkWyuNDvIM4wCgYIKoZIzj0EAwIwDTELMAkGA1UEBhMC
Q0EwIBcNMjAxMTMwMDQxNTU4WhgPMjA3NTA5MDMwNDE1NThaMA0xCzAJBgNVBAYT
AkNBMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESddX8frfPkMZx7wujXUt8lgp
jo1xcDMJqDWYGdaT/YLbPBTykgsozv748+elcPdM96asPskVxh3YaqCBA1ftzaMT
MBEwDwYDVR0RBAgwBocECgAAATAKBggqhkjOPQQDAgNIADBFAiA9OMXMF65y2BGb
1hRmZ4csfm3IGRjEvdGcPoyB8KpmNwIhAN6eiVLgT5To9TRAeHEsl2N+badVIICh
2uHjgoLe3ibY
-----END CERTIFICATE-----`

func TestValidateAccessRequestWithIP(t *testing.T) {
	parsedPrivPEM, _ := crypt.ParseECPrivateKeyPEM(privPEM)
	parsedCertPEMWithIP, _ := getParsedCertificate(certPEMWithIP)

	token := buildToken("10.0.0.1", parsedCertPEMWithIP, parsedPrivPEM)
	result, err := ParseExternalToken(token, createVerifyIssuer("10.0.0.1", HostnameAuth))
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", result)
}

func TestValidateAccessRequestWithWrongIP(t *testing.T) {
	parsedPrivPEM, _ := crypt.ParseECPrivateKeyPEM(privPEM)
	parsedCertPEMWithIP, _ := getParsedCertificate(certPEMWithIP)

	token := buildToken("10.0.0.2", parsedCertPEMWithIP, parsedPrivPEM)
	_, err := ParseExternalToken(token, createVerifyIssuer("10.0.0.2", HostnameAuth))
	assert.EqualError(t, err, "x509: certificate is valid for 10.0.0.1, not 10.0.0.2")
}

const certPEMWrongRoot = `
-----BEGIN CERTIFICATE-----
MIIBNjCB3aADAgECAgkAgMZgCX5kg8YwCgYIKoZIzj0EAwIwDTELMAkGA1UEBhMC
Q0EwHhcNMjAxMTAzMjE0OTExWhcNMjAxMjAzMjE0OTExWjANMQswCQYDVQQGEwJD
QTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABE3DTrT9jjyv0nZCo0LsCv1VeD9/
4+MMQI2L0ZGI5cz+bUJUpBJBaR+O6UAXAf740H3tD13yC5kjb+HhgL5ZCV2jJjAk
MCIGA1UdEQQbMBmCF2NsaWVudC5jbG91ZC5vcHRhYmxlLmNvMAoGCCqGSM49BAMC
A0gAMEUCIQDxmn+B7zBkQnYOtB9f1qlyO5ae9OE2OAZCq8oPoIbAfQIgUMNo2iba
2uNDUjcma9oe2WmuaR5Qr8d+iBad6UBm16M=
-----END CERTIFICATE-----`

const privPEMWrongRoot = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIGPbicHfVZMhQ+v73snf5NpOEOVlT8h9r3IfQXrqn8CpoAoGCCqGSM49
AwEHoUQDQgAETcNOtP2OPK/SdkKjQuwK/VV4P3/j4wxAjYvRkYjlzP5tQlSkEkFp
H47pQBcB/vjQfe0PXfILmSNv4eGAvlkJXQ==
-----END EC PRIVATE KEY-----`

func TestValidateAccessRequestWrongRoot(t *testing.T) {
	parsedPrivPEMWrongRoot, _ := crypt.ParseECPrivateKeyPEM(privPEMWrongRoot)
	parsedCertPEMWrongRoot, _ := getParsedCertificate(certPEMWrongRoot)

	token := buildToken(issuer, parsedCertPEMWrongRoot, parsedPrivPEMWrongRoot)
	_, err := ParseExternalToken(token, createVerifyIssuer(issuer, HostnameAuth))
	assert.Error(t, err, `x509: certificate signed by unknown authority (possibly because of "x509: ECDSA verification failure" while trying to verify candidate authority certificate "serial:13258638361184244459")`)
}
