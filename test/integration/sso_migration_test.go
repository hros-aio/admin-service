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

// setupSSOMigrationTest starts a PostgreSQL container, executes baseline migrations (1 through 4), and returns database references.
func setupSSOMigrationTest(t *testing.T) (*gorm.DB, string, func()) {
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
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000004_create_invite_tokens.up.sql"))

	return db, migDir, cleanup
}

func TestSSOMigrationUpAndDownFlow(t *testing.T) {
	db, migDir, cleanup := setupSSOMigrationTest(t)
	defer cleanup()

	// 1. Run UP migration the first time
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000005_add_sso_to_admin_users.up.sql"))

	// Verify columns exist on admin_users table in active schema
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}
	err := db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'admin_users'").Scan(&columns).Error
	require.NoError(t, err)

	colMap := make(map[string]string)
	for _, col := range columns {
		colMap[col.ColumnName] = col.DataType
	}

	assert.Contains(t, colMap, "sso_identity_id")
	assert.Contains(t, colMap, "sso_provider")

	// Verify uniqueness constraint on sso_identity_id in active schema
	var isUnique bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
			  ON tc.constraint_name = kcu.constraint_name
			  AND tc.table_schema = kcu.table_schema
			WHERE tc.table_schema = 'public'
			  AND tc.table_name = 'admin_users'
			  AND kcu.column_name = 'sso_identity_id'
			  AND tc.constraint_type = 'UNIQUE'
		)
	`).Scan(&isUnique).Error
	require.NoError(t, err)
	assert.True(t, isUnique, "Expected sso_identity_id to have UNIQUE constraint")

	// 2. Run UP migration a second time to verify idempotency
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000005_add_sso_to_admin_users.up.sql"))

	// 3. Run DOWN migration the first time
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000005_add_sso_to_admin_users.down.sql"))

	// Verify columns do not exist on admin_users table in active schema
	err = db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'admin_users'").Scan(&columns).Error
	require.NoError(t, err)

	colMapAfter := make(map[string]string)
	for _, col := range columns {
		colMapAfter[col.ColumnName] = col.DataType
	}

	assert.NotContains(t, colMapAfter, "sso_identity_id")
	assert.NotContains(t, colMapAfter, "sso_provider")

	// Verify uniqueness constraint does not exist on sso_identity_id in active schema
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
			  ON tc.constraint_name = kcu.constraint_name
			  AND tc.table_schema = kcu.table_schema
			WHERE tc.table_schema = 'public'
			  AND tc.table_name = 'admin_users'
			  AND kcu.column_name = 'sso_identity_id'
			  AND tc.constraint_type = 'UNIQUE'
		)
	`).Scan(&isUnique).Error
	require.NoError(t, err)
	assert.False(t, isUnique, "Expected sso_identity_id UNIQUE constraint to be dropped")

	// 4. Run DOWN migration a second time to verify idempotency
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000005_add_sso_to_admin_users.down.sql"))
}
