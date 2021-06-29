package protox

import (
	"fmt"

	v1 "github.com/optable/match-cli/api/v1"
)

type Error struct {
	Res *v1.Error
	Err error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func NewError(code v1.Status, err error) *Error {
	return &Error{
		Res: &v1.Error{Code: code, Message: err.Error()},
		Err: err,
	}
}

func NewNotFoundError(err error) *Error {
	return &Error{
		Res: &v1.Error{Code: v1.Status_STATUS_NOT_FOUND, Message: err.Error()},
		Err: err,
	}
}

func toStringMap(params map[string]interface{}) map[string]string {
	result := make(map[string]string, len(params))
	for key, param := range params {
		if str, ok := param.(fmt.Stringer); ok {
			result[key] = str.String()
		}
	}
	return result
}
