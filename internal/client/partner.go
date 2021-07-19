package client

import (
	"context"
	"fmt"

	v1 "github.com/optable/match-cli/api/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

type PartnerStateNotification int

const (
	UnknownPartnerNotification PartnerStateNotification = iota
	ApprovePartnerNotification
	RejectPartnerNotification
	DisconnectPartnerNotification
)

func (n PartnerStateNotification) String() string {
	return [...]string{"Unknown", "Approve", "Reject", "Disconnect"}[n]
}

func partnerStateNotifyMethod(kind PartnerStateNotification) string {
	switch kind {
	case ApprovePartnerNotification:
		return "approve"
	case RejectPartnerNotification:
		return "reject"
	case DisconnectPartnerNotification:
		return "disconnect"
	case UnknownPartnerNotification:
		fallthrough
	default:
		return ""
	}
}

func (c *AdminRpcClient) CreateInitTokenForPartner(ctx context.Context, req *v1.CreateExternalInitTokenReq) (*v1.PartnersInitToken, error) {
	res := &v1.CreateExternalInitTokenRes{}
	err := c.Do(ctx, "/partners/create", req, res)
	if err != nil {
		return nil, err
	}
	return res.Token, nil
}

func (c *AdminRpcClient) CreatePartner(ctx context.Context, req *v1.ConnectPartnerReq) (*v1.ConnectPartnerRes, error) {
	res := &v1.ConnectPartnerRes{}
	err := c.Do(ctx, "/partners/connect", req, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *AdminRpcClient) ApprovePartner(ctx context.Context, req *v1.ApprovePartnerReq) error {
	return c.Do(ctx, "/partners/approve", req, &emptypb.Empty{})
}

func (c *AdminRpcClient) RejectPartner(ctx context.Context, req *v1.RejectPartnerReq) error {
	return c.Do(ctx, "/partners/reject", req, &emptypb.Empty{})
}

func (c *AdminRpcClient) DisconnectPartner(ctx context.Context, req *v1.DisconnectPartnerReq) error {
	return c.Do(ctx, "/partners/disconnect", req, &emptypb.Empty{})
}

func (c *AdminRpcClient) ListPartners(ctx context.Context) ([]*v1.Partner, error) {
	res := &v1.ListPartnerRes{}
	req := &v1.ListPartnerReq{}

	var partners []*v1.Partner
	for i := 0; ; i++ {
		req.Pagination = &v1.PageReq{Page: int32(i), Size: maxPageSize}
		if err := c.Do(ctx, "/partners/list", req, res); err != nil {
			return nil, err
		}
		partners = append(partners, res.GetData()...)
		if !res.GetPagination().GetHasMore() {
			break
		}
	}

	return partners, nil
}

func (c *AdminRpcClient) ConnectPartner(ctx context.Context, req *v1.ExternalInfo) error {
	res := &emptypb.Empty{}
	return c.Do(ctx, "/partner/connect", req, res)
}

func (c *AdminRpcClient) NotifyPartner(ctx context.Context, notification PartnerStateNotification) error {
	method := partnerStateNotifyMethod(notification)
	if method == "" {
		return fmt.Errorf("unsupported partner state notification \"%s\"", notification)
	}
	req := &emptypb.Empty{}
	res := &emptypb.Empty{}
	return c.Do(ctx, "/partner/"+method, req, res)
}

func (c *AdminRpcClient) RegisterPartner(ctx context.Context, req *v1.RegisterExternalPartnerReq) error {
	res := &emptypb.Empty{}
	return c.Do(ctx, "/partner/register", req, res)
}
