package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) WithinTx(ctx context.Context, opts pgx.TxOptions, fn func(*Repositories) error) error {
	tx, err := m.pool.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	repositories := NewRepositories(tx)

	if err := fn(repositories); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
