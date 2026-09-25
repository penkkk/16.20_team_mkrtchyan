package service

import (
	"opd/internal/repository"
	authservice "opd/internal/service/auth"
	token "opd/internal/service/token"
	// libraryservice "opd/internal/service/library"
)

type Services struct {
	Auth authservice.Service
	// Library libraryservice.Service
}

func NewServices(repositories *repository.Repositories, tokens token.Manager) *Services {
	return &Services{
		Auth: authservice.NewService(repositories.Auth, tokens),
	}
}
