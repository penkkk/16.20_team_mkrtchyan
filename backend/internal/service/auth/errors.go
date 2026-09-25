package auth

import (
	"errors"
	"strings"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotImplemented     = errors.New("not implemented")
)

type FieldConflictError struct {
	Fields []string
}

func NewFieldConflictError(fields []string) *FieldConflictError {
	return &FieldConflictError{
		Fields: fields,
	}
}

func (e *FieldConflictError) Error() string {
	return "fields already taken: " + strings.Join(e.Fields, ", ")
}
