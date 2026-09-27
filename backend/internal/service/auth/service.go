package auth

import (
	"context"

	"opd/internal/repository"
	authrepo "opd/internal/repository/auth"
	token "opd/internal/service/token"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type AuthService interface {
	Login(ctx context.Context, input LoginInput) (LoginResult, error)
	Register(ctx context.Context, input RegisterInput) (LoginResult, error)
	Logout(ctx context.Context, input LogoutInput) error
	Refresh(ctx context.Context, input LogoutInput) (RefreshResult, error)
}

type authRepository interface {
	CreateUser(ctx context.Context, input authrepo.CreateUserInput) (authrepo.User, error)
	CreateUserPassword(ctx context.Context, input authrepo.CreateUserPasswordInput) (authrepo.UserPassword, error)
	GetUserCredentialsByLogin(ctx context.Context, login string) (authrepo.UserCredentials, error)
	FindUserConflicts(ctx context.Context, input authrepo.FindUserConflictsInput) ([]string, error)
	CreateRefreshSession(ctx context.Context, input authrepo.CreateRefreshSessionInput) (authrepo.RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) (authrepo.RevokedRefreshSession, error)
}

type authTxManager interface {
	WithinTx(ctx context.Context, opts pgx.TxOptions, fn func(*repository.Repositories) error) error
}

type authService struct {
	repo        authRepository
	txManager   authTxManager
	tokens      token.Manager
	redisClient *redis.Client
}

func NewAuthService(
	repo authRepository,
	txManager authTxManager,
	tokens token.Manager,
	redisClient *redis.Client,
) AuthService {
	return &authService{
		repo:        repo,
		txManager:   txManager,
		tokens:      tokens,
		redisClient: redisClient,
	}
}
