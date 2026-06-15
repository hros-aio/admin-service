package database

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGormTxManager_WithinTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	txManager := NewTxManager(gormDB)

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectCommit()

		err := txManager.WithinTx(context.Background(), func(_ context.Context) error {
			tx := GetTx(context.Background(), gormDB)
			assert.NotNil(t, tx)
			return nil
		})

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectRollback()

		err := txManager.WithinTx(context.Background(), func(_ context.Context) error {
			return gorm.ErrInvalidData
		})

		require.ErrorIs(t, err, gorm.ErrInvalidData)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
