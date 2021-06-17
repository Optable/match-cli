package auth

import (
	"net/http"
)

type AuthStrategy interface {
	Authenticate(req *http.Request) (interface{}, error)
}

type AuthStrategyFn func(req *http.Request) (interface{}, error)

func (fn AuthStrategyFn) Authenticate(req *http.Request) (interface{}, error) {
	return fn(req)
}

func NewChainedAuthStrategy(strategies ...AuthStrategy) AuthStrategy {
	return AuthStrategyFn(func(req *http.Request) (resource interface{}, err error) {
		// Try all strategies until we find a successful one
		for _, strategy := range strategies {
			resource, err = strategy.Authenticate(req)
			if err == nil {
				return resource, nil
			}
		}

		return nil, err
	})
}
