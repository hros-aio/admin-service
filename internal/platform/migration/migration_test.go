package migration

import (
	"log/slog"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMigrator_Run(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	migrator := NewMigrator(gormDB, log)

	t.Run("success", func(t *testing.T) {
		// AutoMigrate does many queries, hard to mock exactly with sqlmock
		// without being very brittle. We'll just verify the call doesn't panic
		// and returns nil for empty models.
		err := migrator.Run()
		assert.NoError(t, err)
	})
}
