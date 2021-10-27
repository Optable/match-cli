package client

import (
	"context"

	v1 "github.com/optable/match-api/match/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (c *AdminRpcClient) CreateInitTokenForPartner(ctx context.Context, req *v1.CreateExternalInitTokenReq) (*v1.PartnersInitToken, error) {
	res := &v1.CreateExternalInitTokenRes{}
	err := c.Do(ctx, "/partners/create", req, res)
	if err != nil {
		return nil, err
	}
	return res.Token, nil
}

func (c *AdminRpcClient) RegisterPartner(ctx context.Context, req *v1.RegisterExternalPartnerReq) error {
	res := &emptypb.Empty{}
	return c.Do(ctx, "/partner/register", req, res)
}
