package db

import (
	"context"
	"database/sql"

	"money-buddy-backend/internal/services"
)

type SQLTxManager struct {
	db *sql.DB
}

func NewSQLTxManager(db *sql.DB) *SQLTxManager {
	return &SQLTxManager{db: db}
}

func (m *SQLTxManager) Begin(ctx context.Context) (services.Tx, error) {
	return m.db.BeginTx(ctx, nil)
}
