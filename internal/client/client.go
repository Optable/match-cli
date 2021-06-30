package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	v1 "github.com/optable/match-cli/api/v1"
	"github.com/optable/match-cli/internal/auth"
	"github.com/optable/match-cli/internal/protox"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const maxPageSize = 100

type AdminRpcClient struct {
	*http.Client
	url         string
	tokenSource TokenSource
}

type TokenSource interface {
	Token(req *http.Request) (string, error)
}

type TokenSourceFn func(req *http.Request) (string, error)

func (fn TokenSourceFn) Token(req *http.Request) (string, error) {
	return fn(req)
}

func StaticTokenSource(authToken string) TokenSource {
	return TokenSourceFn(func(_ *http.Request) (string, error) {
		return authToken, nil
	})
}

func UserCredentialsTokenSource(email string, secret string, authAPIURL string) TokenSource {
	return TokenSourceFn(func(req *http.Request) (string, error) {
		res, err := CreateUserSession(req.Context(), authAPIURL, &v1.CreateSessionReq{
			Email:  email,
			Secret: secret,
		})

		if err != nil {
			return "", err
		}

		return res.Token, nil
	})
}

func ServiceKeyTokenSource(keyID string, privateKeyDer string) TokenSource {
	return TokenSourceFn(func(req *http.Request) (string, error) {
		hasher := sha256.New()
		body, err := req.GetBody()
		if err != nil {
			return "", err
		}

		bytes, err := ioutil.ReadAll(body)
		if err != nil {
			return "", err
		}
		checksum := hasher.Sum(bytes)

		token := jwt.NewWithClaims(jwt.SigningMethodES256, &auth.SignedRequestClaims{
			StandardClaims: jwt.StandardClaims{Issuer: keyID},
			B:              base64.URLEncoding.EncodeToString(checksum),
			TS:             time.Now().Unix(),
		})

		decoded, err := base64.StdEncoding.DecodeString(privateKeyDer)
		if err != nil {
			return "", err
		}

		privateKey, err := x509.ParseECPrivateKey(decoded)
		if err != nil {
			return "", err
		}

		signedToken, err := token.SignedString(privateKey)
		if err != nil {
			return "", err
		}

		return signedToken, nil
	})
}

func NewClient(url string, tokenSource TokenSource, client *http.Client) *AdminRpcClient {
	if client == nil {
		client = &http.Client{}
	}

	// Remove trailing slashes
	url = strings.TrimRight(url, "/")

	return &AdminRpcClient{Client: client, url: url, tokenSource: tokenSource}
}

// Implementation details

func (c *AdminRpcClient) path(method string) string {
	return c.url + method
}

func (c *AdminRpcClient) Do(ctx context.Context, method string, req, res proto.Message) error {
	var httpReqMethod string
	if req != nil {
		httpReqMethod = "POST"
	} else {
		httpReqMethod = "GET"
	}

	msg, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, httpReqMethod, c.path(method), bytes.NewBuffer(msg))
	if err != nil {
		return err
	}

	if c.tokenSource != nil {
		token, err := c.tokenSource.Token(httpReq)
		if err != nil {
			return err
		}
		httpReq.Header.Add("Authorization", "Bearer "+token)
	}
	httpReq.Header.Add("Content-Type", "application/protobuf")

	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode != http.StatusOK {
		res := &v1.Error{}

		respErr := fmt.Errorf(
			"Unexpected status code for %s %s: %s",
			httpReqMethod, c.path(method), httpResp.Status,
		)

		if err := proto.Unmarshal(body, res); err != nil {
			return fmt.Errorf("Error without body: %w", respErr)
		}

		errString, err := protojson.Marshal(res)
		if err != nil {
			return err
		}

		return &protox.Error{
			Res: res,
			Err: fmt.Errorf(respErr.Error()+": %s", errString),
		}
	}

	return proto.Unmarshal(body, res)
}
