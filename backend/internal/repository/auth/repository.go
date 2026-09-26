package auth

import (
	"opd/internal/db"
)

type AuthPostgresRepository struct {
	queries *db.Queries
}

func NewAuthPostgresRepository(dbt db.DBTX) *AuthPostgresRepository {
	return &AuthPostgresRepository{
		queries: db.New(dbt),
	}
}
