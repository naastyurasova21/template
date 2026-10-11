package txmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/naastyurasova21/template/internal/db"
)

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type manager struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) TxManager {
	return &manager{pool: pool}
}

const rollbackTimeout = 5 * time.Second

func (m *manager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := db.TxFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	rollbackCtx, cancelRollback := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
	defer cancelRollback()

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(rollbackCtx)
			panic(p)
		}
	}()

	txCtx := db.WithTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(rollbackCtx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
