package auth

import (
	"context"

	"opd/internal/db"
)

type Repository interface {
	CreateUser(ctx context.Context, input CreateUserInput) (User, error)
	CreateUserPassword(ctx context.Context, input CreateUserPasswordInput) (UserPassword, error)
	GetUserCredentialsByLogin(ctx context.Context, login string) (UserCredentials, error)
	CreateRefreshSession(ctx context.Context, input CreateRefreshSessionInput) (RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) (RefreshSession, error)
}

type PostgresRepository struct {
	queries *db.Queries
}

func NewPostgresRepository(dbt db.DBTX) *PostgresRepository {
	return &PostgresRepository{
		queries: db.New(dbt),
	}
}

func NewRepository(dbt db.DBTX) Repository {
	return NewPostgresRepository(dbt)
}
