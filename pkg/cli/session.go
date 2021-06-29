package cli

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	v1 "github.com/optable/match-cli/api/v1"
	"github.com/optable/match-cli/pkg/auth"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CreateUserSession(ctx context.Context, authAPIURL string, req *v1.CreateSessionReq) (*v1.CreateSessionRes, error) {
	body, err := protojson.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, authAPIURL+"/sign-in", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer httpRes.Body.Close()

	for _, c := range httpRes.Cookies() {
		if c.Name == auth.UserAccessCookie {
			return &v1.CreateSessionRes{
				Expiry: timestamppb.New(c.Expires),
				Token:  c.Value,
			}, nil
		}
	}

	return nil, errors.New("Couldn't read access token from sign in response")
}
