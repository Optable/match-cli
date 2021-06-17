package auth

import (
	v1 "github.com/optable/match-cli/api/v1"
)

type AccessScope = string

const (
	AdminOwnerScope AccessScope = "admin.owner"
)

var AvailableScopes = []string{AdminOwnerScope}

var InternalAccount = v1.Account{
	Name:         "INTERNAL",
	AccessScopes: AvailableScopes,
	Kind:         v1.AccountKind_ACCOUNT_KIND_SERVICE,
}

func HasAccessScope(a *v1.Account, scope string) bool {
	for _, s := range a.GetAccessScopes() {
		if s == scope {
			return true
		}
	}

	return false
}
