package repository

import (
	"opd/internal/db"
	authrepo "opd/internal/repository/auth"
	bookrepo "opd/internal/repository/book"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	Auth      *authrepo.AuthPostgresRepository
	Book      *bookrepo.PostgresRepository
	TxManager *TxManager
}

func NewRepositories(dbt db.DBTX) *Repositories {
	return &Repositories{
		Auth: authrepo.NewAuthPostgresRepository(dbt),
		Book: bookrepo.NewPostgresRepository(dbt),
	}
}

func NewRepositoriesWithTxManager(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Auth:      authrepo.NewAuthPostgresRepository(pool),
		Book:      bookrepo.NewPostgresRepository(pool),
		TxManager: NewTxManager(pool),
	}
}
