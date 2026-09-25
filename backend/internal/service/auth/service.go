package auth

import (
	"context"
	"opd/internal/repository"
	authrepo "opd/internal/repository/auth"
	token "opd/internal/service/token"

	"github.com/jackc/pgx/v5"
)

type AuthService interface {
	Login(ctx context.Context, input LoginInput) (LoginResult, error)
	Register(ctx context.Context, input RegisterInput) (LoginResult, error)
}

type authRepository interface {
	CreateUser(ctx context.Context, input authrepo.CreateUserInput) (authrepo.User, error)
	CreateUserPassword(ctx context.Context, input authrepo.CreateUserPasswordInput) (authrepo.UserPassword, error)
	GetUserCredentialsByLogin(ctx context.Context, login string) (authrepo.UserCredentials, error)
	FindUserConflicts(ctx context.Context, input authrepo.FindUserConflictsInput) ([]string, error)
	CreateRefreshSession(ctx context.Context, input authrepo.CreateRefreshSessionInput) (authrepo.RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) (authrepo.RefreshSession, error)
}

type authTxManager interface {
	WithinTx(ctx context.Context, opts pgx.TxOptions, fn func(*repository.Repositories) error) error
}

type authService struct {
	repo      authRepository
	txManager authTxManager
	tokens    token.Manager
}

func NewAuthService(repo authRepository, txManager authTxManager, tokens token.Manager) AuthService {
	return &authService{
		repo:      repo,
		txManager: txManager,
		tokens:    tokens,
	}
}
