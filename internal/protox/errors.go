package protox

import (
	v1 "github.com/optable/match-api/match/v1"
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
