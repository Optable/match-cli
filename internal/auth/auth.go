package auth

import (
	"github.com/dgrijalva/jwt-go"
)

const UserAccessCookie = "OPTABLE_ADMIN_ACCESS"

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
