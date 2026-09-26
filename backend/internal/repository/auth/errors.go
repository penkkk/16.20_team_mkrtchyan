package auth

import (
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found")

type UniqueConstraintError struct {
	Fields []string
}

func NewUniqueConstraintError(fields ...string) *UniqueConstraintError {
	return &UniqueConstraintError{
		Fields: fields,
	}
}

func (e *UniqueConstraintError) Error() string {
	return "unique constraint violation: " + strings.Join(e.Fields, ", ")
}
