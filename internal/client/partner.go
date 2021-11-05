package client

import (
	"context"

	v1 "github.com/optable/match-api/match/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (c *OptableRpcClient) RegisterPartner(ctx context.Context, req *v1.RegisterPartnerReq) error {
	res := &emptypb.Empty{}
	return c.Do(ctx, "/partner/register", req, res)
}
