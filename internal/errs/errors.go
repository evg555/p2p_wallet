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
	ErrWalletNotFound     = errors.New("wallet not found")
	ErrWalletMismatch     = errors.New("wallet doesn't belong to current user")
	ErrNotEnoughMoney     = errors.New("not enough money for transfer")
	ErrCurrencyMismatch   = errors.New("currencies mismatch within transaction")
	ErrNotPositiveAmount  = errors.New("amount must be positive")
)

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}
