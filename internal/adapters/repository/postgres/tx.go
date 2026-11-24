package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DBTX abstrai *sqlx.DB e *sqlx.Tx para que os repositórios participem transparentemente de transações.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type txContextKey struct{}

// TxManager gerencia transações PostgreSQL usando *sqlx.DB.
type TxManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if existing := txFromContext(ctx); existing != nil {
		return fn(ctx)
	}

	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// WithoutTx remove uma transação ativa do contexto para operações que precisam persistir mesmo se a tx atual sofrer rollback.
func WithoutTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, txContextKey{}, nil)
}

func txFromContext(ctx context.Context) *sqlx.Tx {
	if ctx == nil {
		return nil
	}
	tx, _ := ctx.Value(txContextKey{}).(*sqlx.Tx)
	return tx
}

func getExecutor(ctx context.Context, db *sqlx.DB) DBTX {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return db
}
