package auth

import (
	"errors"
	"strings"
)

var (
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrNotImplemented            = errors.New("not implemented")
	ErrInvalidRefreshSession     = errors.New("invalid refresh session")
	ErrInvalidAccessToken        = errors.New("invalid access token")
	ErrUnsupportedOAuthProvider  = errors.New("unsupported oauth provider")
	ErrOAuthClientIDRequired     = errors.New("oauth client id is required")
	ErrOAuthClientSecretRequired = errors.New("oauth client secret is required")
	ErrOAuthRedirectURIRequired  = errors.New("oauth redirect uri is required")
	ErrOAuthProviderMismatch     = errors.New("oauth provider mismatch")
	ErrOAuthStateInvalid         = errors.New("oauth state invalid")
	ErrOAuthCodeRequired         = errors.New("oauth code required")
	ErrOAuthNonceMismatch        = errors.New("oauth nonce mismatch")
	ErrOAuthEmailAlreadyUsed     = errors.New("oauth email already used")
	ErrOAuthUsernameRequired     = errors.New("oauth username required")
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
