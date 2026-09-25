package service

import (
	"opd/internal/repository"
	authservice "opd/internal/service/auth"
	token "opd/internal/service/token"
)

type Services struct {
	Auth authservice.AuthService
}

func NewServices(repositories *repository.Repositories, tokens token.Manager) *Services {
	return &Services{
		Auth: authservice.NewAuthService(repositories.Auth, repositories.TxManager, tokens),
	}
}
