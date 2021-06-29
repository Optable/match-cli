package auth

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"optable-sandbox/pkg/service/admin/protox"

	"github.com/optable/match-cli/api/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

func NewUserStrategy(exec boil.ContextExecutor, userAccessSecret string) AuthStrategy {
	return AuthStrategyFn(func(req *http.Request) (interface{}, error) {
		var token string
		cookie, cerr := req.Cookie(UserAccessCookie)
		if cerr == nil && cookie.Value != "" {
			token = cookie.Value
		}

		// Accept user access token via auth header as well
		if token == "" {
			token = getHeaderToken(req)
		}

		if token == "" {
			return nil, errors.New("user: missing access token")
		}

		parser := jwt.Parser{ValidMethods: []string{jwt.SigningMethodHS256.Alg()}}

		authToken, err := parser.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(userAccessSecret), nil
		})

		if err != nil {
			return nil, err
		}

		claims, ok := authToken.Claims.(*jwt.StandardClaims)
		if !ok || !authToken.Valid {
			return nil, errors.New("user: invalid access token")
		}

		userID, err := strconv.ParseInt(claims.Subject, 10, 32)
		if err != nil {
			return nil, errors.New("user: invalid subject")
		}

		account, err := models.Accounts(
			models.AccountWhere.Kind.EQ(models.AccountKindEnumUser),
			models.AccountWhere.ID.EQ(int(userID)),
			models.AccountWhere.Status.EQ(models.AccountsStatusEnumActive),
		).One(req.Context(), exec)
		if err != nil {
			return nil, err
		}

		return protox.AccountFromRecord(account), nil
	})
}

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

func NewServiceAccountAuthStrategy(db *sql.DB) AuthStrategy {
	return AuthStrategyFn(func(req *http.Request) (interface{}, error) {
		var key *models.ServiceAccountKey

		_, err := ParseSignedRequestToken(req, func(issuer string) (crypto.PublicKey, error) {
			var err error

			key, err = models.ServiceAccountKeys(
				qm.Expr(
					models.ServiceAccountKeyWhere.ExpiresAt.IsNull(),
					qm.Or2(models.ServiceAccountKeyWhere.ExpiresAt.GT(null.NewTime(time.Now(), true))),
				),
				models.ServiceAccountKeyWhere.ID.EQ(issuer),
			).One(req.Context(), db)

			if err != nil {
				return nil, err
			}

			activeAccountMod := models.AccountWhere.Status.EQ(models.AccountsStatusEnumActive)
			if _, err = key.Account(activeAccountMod).One(req.Context(), db); err != nil {
				return nil, err
			}

			decodedPublicKey, err := base64.StdEncoding.DecodeString(key.PublicKeyDer)
			if err != nil {
				return nil, err
			}
			unmarshalledPublicKey, err := x509.ParsePKIXPublicKey(decodedPublicKey)
			if err != nil {
				return nil, err
			}

			return unmarshalledPublicKey, nil
		})

		if err != nil {
			return nil, fmt.Errorf("service: parse signed request token: %w", err)
		}

		account, err := key.Account().One(req.Context(), db)
		if err != nil {
			return nil, fmt.Errorf("service: load account: %w", err)
		}

		return protox.AccountFromRecord(account), nil
	})
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
