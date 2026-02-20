package errs

import (
	"errors"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrPasswordMismatch = errors.New("password mismatch")
)
