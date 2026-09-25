package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotImplemented     = errors.New("not implemented")
)
