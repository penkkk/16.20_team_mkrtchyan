package book

import "opd/internal/db"

type PostgresRepository struct {
	db db.DBTX
}

func NewPostgresRepository(dbt db.DBTX) *PostgresRepository {
	return &PostgresRepository{db: dbt}
}
