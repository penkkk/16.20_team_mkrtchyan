package repository

import (
	"opd/internal/db"
	authrepo "opd/internal/repository/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	Auth      *authrepo.AuthPostgresRepository
	TxManager *TxManager
}

func NewRepositories(dbt db.DBTX) *Repositories {
	return &Repositories{
		Auth: authrepo.NewAuthPostgresRepository(dbt),
	}
}

func NewRepositoriesWithTxManager(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Auth:      authrepo.NewAuthPostgresRepository(pool),
		TxManager: NewTxManager(pool),
	}
}
