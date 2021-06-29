package auth

import (
	"context"

	v1 "github.com/optable/match-cli/api/v1"
)

type key int

const (
	contextAccount key = iota
	contextPartner
)

// AccountFromContext allows getting the current account from request context
func AccountFromContext(ctx context.Context) *v1.Account {
	acc, _ := ctx.Value(contextAccount).(*v1.Account)
	return acc
}

// ContextWithAccount allows to wrap a given context with account info
func ContextWithAccount(ctx context.Context, account *v1.Account) context.Context {
	return context.WithValue(ctx, contextAccount, account)
}

func PartnerFromContext(ctx context.Context) string {
	acc, _ := ctx.Value(contextPartner).(string)
	return acc
}

func ContextWithPartner(ctx context.Context, partner string) context.Context {
	return context.WithValue(ctx, contextPartner, partner)
}
