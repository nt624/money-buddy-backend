package db

import (
	"database/sql"

	"money-buddy-backend/infra/transaction"
	"money-buddy-backend/internal/services"
)

// NewSQLTxManager は TxManager の infra 実装を返します。
func NewSQLTxManager(db *sql.DB) services.TxManager {
	return transaction.NewTxManager(db)
}
