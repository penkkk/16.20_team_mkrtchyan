package auth

import (
	"context"
	authrepo "opd/internal/repository/auth"
	token "opd/internal/service/token"
)

type Service interface {
	Login(ctx context.Context, input LoginInput) (LoginResult, error)
	Register(ctx context.Context, input RegisterInput) error
}

type service struct {
	repo   authrepo.Repository
	tokens token.Manager
}

func NewService(repo authrepo.Repository, tokens token.Manager) Service {
	return &service{
		repo:   repo,
		tokens: tokens,
	}
}
