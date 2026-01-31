package transaction

import (
	"context"
	"database/sql"

	"money-buddy-backend/internal/services"
)

type txKey struct{}

func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}

type sqlTx struct {
	tx *sql.Tx
}

func (t *sqlTx) Commit() error {
	return t.tx.Commit()
}

func (t *sqlTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *sqlTx) Context(ctx context.Context) context.Context {
	return withTx(ctx, t.tx)
}

type sqlTxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) services.TxManager {
	return &sqlTxManager{db: db}
}

func (m *sqlTxManager) Begin(ctx context.Context) (services.Tx, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqlTx{tx: tx}, nil
}
