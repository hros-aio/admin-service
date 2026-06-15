package database

import (
	"context"

	"gorm.io/gorm"
)

// TxManager defines the interface for managing database transactions.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type txKey struct{}

// GormTxManager is the GORM implementation of TxManager.
type GormTxManager struct {
	db *gorm.DB
}

// NewTxManager creates a new GormTxManager.
func NewTxManager(db *gorm.DB) TxManager {
	return &GormTxManager{db: db}
}

// WithinTx executes the given function within a database transaction.
func (m *GormTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

// GetTx retrieves the GORM transaction from the context if it exists.
func GetTx(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return defaultDB.WithContext(ctx)
}
