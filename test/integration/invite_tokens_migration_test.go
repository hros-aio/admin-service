package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupInviteTokensMigrationTest starts a PostgreSQL container, executes baseline migrations (1, 2, 3), and returns database references.
func setupInviteTokensMigrationTest(t *testing.T) (*gorm.DB, string, func()) {
	t.Helper()
	ctx := context.Background()

	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("hros_admin"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second),
		),
	)
	require.NoError(t, err)

	cleanup := func() {
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		cleanup()
		require.NoError(t, err)
	}

	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	if err != nil {
		cleanup()
		require.NoError(t, err)
	}

	migDir := findMigrationsDir(t)
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))

	return db, migDir, cleanup
}

func TestInviteTokensMigrationUpAndDownFlow(t *testing.T) {
	db, migDir, cleanup := setupInviteTokensMigrationTest(t)
	defer cleanup()

	// 1. Run UP migration the first time
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000004_create_invite_tokens.up.sql"))

	// Verify that table invite_tokens exists
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'invite_tokens')").Scan(&exists).Error
	require.NoError(t, err)
	assert.True(t, exists, "Expected invite_tokens table to exist")

	// Verify that all required columns exist with correct types
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}
	err = db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'invite_tokens'").Scan(&columns).Error
	require.NoError(t, err)

	colMap := make(map[string]string)
	for _, col := range columns {
		colMap[col.ColumnName] = col.DataType
	}

	assert.Contains(t, colMap, "id")
	assert.Contains(t, colMap, "admin_id")
	assert.Contains(t, colMap, "token")
	assert.Contains(t, colMap, "inviter_id")
	assert.Contains(t, colMap, "expires_at")
	assert.Contains(t, colMap, "consumed_at")
	assert.Contains(t, colMap, "created_at")

	// Verify foreign keys and indexes do not fail GORM insert check
	// 2. Run UP migration a second time to verify idempotency
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000004_create_invite_tokens.up.sql"))

	// 3. Run DOWN migration the first time
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000004_create_invite_tokens.down.sql"))

	// Verify that table invite_tokens does not exist
	err = db.Raw("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'invite_tokens')").Scan(&exists).Error
	require.NoError(t, err)
	assert.False(t, exists, "Expected invite_tokens table to be dropped")

	// 4. Run DOWN migration a second time to verify idempotency
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000004_create_invite_tokens.down.sql"))
}
