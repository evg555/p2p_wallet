package errs

import (
	"errors"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrPasswordMismatch   = errors.New("password mismatch")
	ErrAccessDenied       = errors.New("access denied")
	ErrUserAlreadyExist   = errors.New("user already exist")
	ErrSessionNotFound    = errors.New("session not found")
	ErrEmptyField         = errors.New("field is empty")
	ErrFieldTooMuch       = errors.New("field is too much")
	ErrWalletAlreadyExist = errors.New("wallet already exist")
)

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}
