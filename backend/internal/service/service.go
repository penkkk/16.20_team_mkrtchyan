package service

import (
	"opd/internal/repository"
	authservice "opd/internal/service/auth"
	token "opd/internal/service/token"

	"github.com/redis/go-redis/v9"
)

type Services struct {
	Auth authservice.AuthService
}

func NewServices(
	repositories *repository.Repositories,
	tokens token.Manager,
	redisClient *redis.Client,
) *Services {
	return &Services{
		Auth: authservice.NewAuthService(
			repositories.Auth,
			repositories.TxManager,
			tokens,
			redisClient,
		),
	}
}
