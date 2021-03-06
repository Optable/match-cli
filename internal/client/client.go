package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	v1 "github.com/optable/match-api/match/v1"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type OptableRpcClient struct {
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

func NewClient(url string, tokenSource TokenSource) *OptableRpcClient {
	client := &http.Client{}

	// Remove trailing slashes
	url = strings.TrimRight(url, "/")

	return &OptableRpcClient{Client: client, url: url, tokenSource: tokenSource}
}

// Implementation details

func (c *OptableRpcClient) path(method string) string {
	return c.url + method
}

func (c *OptableRpcClient) Do(ctx context.Context, method string, req, res proto.Message) error {
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
			"unexpected status code for %s %s: %s",
			httpReqMethod, c.path(method), httpResp.Status,
		)

		if err := proto.Unmarshal(body, res); err != nil {
			return fmt.Errorf("error without body: %w", respErr)
		}

		errString, err := protojson.Marshal(res)
		if err != nil {
			return err
		}

		return fmt.Errorf(respErr.Error()+": %s", errString)
	}

	return proto.Unmarshal(body, res)
}
