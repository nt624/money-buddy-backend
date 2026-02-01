package services

import "context"

// Tx はトランザクションの最小インターフェースです。
type Tx interface {
	Commit() error
	Rollback() error
	Context(ctx context.Context) context.Context
}

// TxManager はトランザクション開始を担うインターフェースです。
type TxManager interface {
	Begin(ctx context.Context) (Tx, error)
}
