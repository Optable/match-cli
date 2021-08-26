package client

import (
	"context"

	v1 "github.com/optable/match-cli/api/v1"
)

func (c *AdminRpcClient) CreateMatch(ctx context.Context, req *v1.CreateExternalMatchReq) (*v1.CreateExternalMatchRes, error) {
	res := &v1.CreateExternalMatchRes{}
	if err := c.Do(ctx, "/match/create", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *AdminRpcClient) RunMatch(ctx context.Context, req *v1.RunExternalMatchReq) (*v1.RunExternalMatchRes, error) {
	res := &v1.RunExternalMatchRes{}
	if err := c.Do(ctx, "/match/run", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *AdminRpcClient) GetResult(ctx context.Context, req *v1.GetExternalMatchResultReq) (*v1.GetExternalMatchResultRes, error) {
	res := &v1.GetExternalMatchResultRes{}
	if err := c.Do(ctx, "/match/get-result", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *AdminRpcClient) ListMatches(ctx context.Context, req *v1.ListExternalMatchReq) (*v1.ListExternalMatchRes, error) {
	res := &v1.ListExternalMatchRes{}
	if err := c.Do(ctx, "/match/list", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *AdminRpcClient) GetMatchResults(ctx context.Context, req *v1.GetExternalMatchResultsReq) (*v1.GetExternalMatchResultsRes, error) {
	res := &v1.GetExternalMatchResultsRes{}
	if err := c.Do(ctx, "/match/get-results", req, res); err != nil {
		return nil, err
	}
	return res, nil
}
