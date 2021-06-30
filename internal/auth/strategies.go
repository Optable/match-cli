package auth

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var tokenSplitter = regexp.MustCompile("(?i)bearer")

func getHeaderToken(r *http.Request) string {
	headerParts := tokenSplitter.Split(r.Header.Get("Authorization"), 2)
	if len(headerParts) != 2 {
		return ""
	}

	return strings.TrimSpace(headerParts[1])
}

const UserAccessCookie = "OPTABLE_ADMIN_ACCESS"

func NewInternalAuthStrategy(internalAuthToken string) AuthStrategy {
	return AuthStrategyFn(func(req *http.Request) (interface{}, error) {
		token := getHeaderToken(req)
		if token == "" {
			return nil, errors.New("internal: missing access token")
		}

		if token != internalAuthToken {
			return nil, errors.New("internal: invalid access token")
		}
		return &InternalAccount, nil
	})
}

type SignedRequestClaims struct {
	jwt.StandardClaims
	// Follow draft suggestion for signed requests
	// https://tools.ietf.org/html/draft-richanna-http-jwt-signature-00#section-3.1
	// b field includes a url base64 encoded SHA256 checksum of the request body
	B string `json:"b"`
	// ts field includes a unix timestamp of when the request was generated
	// It's server responsibility to evaluate staleness
	// It's preferable to exp claim which leave too much responsibility on the client.
	// It's there to mitigate replay attacks
	TS int64
}

var SignedRequestMaxAge = 10 * time.Minute

func ParseSignedRequestToken(req *http.Request, findKey func(issuer string) (crypto.PublicKey, error)) (*SignedRequestClaims, error) {
	token := getHeaderToken(req)

	if token == "" {
		return nil, errors.New("missing auth token")
	}

	claims := &SignedRequestClaims{}
	hasher := sha256.New()
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

	checksum := base64.URLEncoding.EncodeToString(hasher.Sum(bodyBytes))

	parser := jwt.Parser{ValidMethods: []string{jwt.SigningMethodES256.Alg()}}
	parsedToken, err := parser.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if claims.B == "" || claims.Issuer == "" || claims.TS == 0 {
			return nil, errors.New("missing issuer or checksum claims")
		}

		if claims.B != checksum {
			return nil, errors.New("checksum claim doesn't match request body")
		}

		if time.Since(time.Unix(claims.TS, 0)) > SignedRequestMaxAge {
			return nil, errors.New("timestamp claim is invalid, request too old/new")
		}

		return findKey(claims.Issuer)
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
