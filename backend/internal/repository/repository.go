package repository

import (
	"opd/internal/db"
	authrepo "opd/internal/repository/auth"
)

type Repositories struct {
	Auth authrepo.Repository
}

func NewRepositories(dbt db.DBTX) *Repositories {
	return &Repositories{
		Auth: authrepo.NewRepository(dbt),
	}
}
