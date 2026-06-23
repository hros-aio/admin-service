package integration

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMFAMigrationFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Setup testcontainers PostgreSQL instance
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("hros_admin"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Connect directly with GORM to run migrations and seed test data
	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// 2. Execute base migrations
	migDir := findMigrationsDir(t)
	runSQLFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runSQLFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))

	// 3. Seed roles and admin user with old MFA secret format
	roleID := domain.NewUUID()
	err = db.Exec("INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		roleID, "Super Admin", "Super Admin Role", true).Error
	require.NoError(t, err)

	adminUserID := domain.NewUUID()
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status, mfa_enabled, mfa_secret) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		adminUserID, "MFA Test Admin", "mfa-admin@hros.com", "hash", roleID, "active", true, "my_super_secret_totp").Error
	require.NoError(t, err)

	// Verify columns exist initially
	var mfaSecret string
	err = db.Raw("SELECT mfa_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&mfaSecret).Error
	require.NoError(t, err)
	assert.Equal(t, "my_super_secret_totp", mfaSecret)

	// 4. Run up migration: 000003_add_mfa_to_admin_users.up.sql
	runSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))

	// Verify that mfa_secret column is dropped (selecting it should error)
	var dummy string
	err = db.Raw("SELECT mfa_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "mfa_secret") || strings.Contains(err.Error(), "does not exist"), "Expected error about mfa_secret column not existing, got: %v", err)

	// Verify that totp_secret column has the copied data
	var totpSecret string
	err = db.Raw("SELECT totp_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&totpSecret).Error
	require.NoError(t, err)
	assert.Equal(t, "my_super_secret_totp", totpSecret)

	// Verify that webauthn_credentials column exists and is null
	var webauthnCreds *string
	err = db.Raw("SELECT webauthn_credentials FROM admin_users WHERE id = ?", adminUserID).Scan(&webauthnCreds).Error
	require.NoError(t, err)
	assert.Nil(t, webauthnCreds)

	// 5. Run down migration: 000003_add_mfa_to_admin_users.down.sql
	runSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.down.sql"))

	// Verify totp_secret column is dropped (selecting it should error)
	err = db.Raw("SELECT totp_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	assert.Error(t, err)

	// Verify webauthn_credentials column is dropped (selecting it should error)
	err = db.Raw("SELECT webauthn_credentials FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	assert.Error(t, err)

	// Verify that mfa_secret column is restored with correct data
	var restoredMfaSecret string
	err = db.Raw("SELECT mfa_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&restoredMfaSecret).Error
	require.NoError(t, err)
	assert.Equal(t, "my_super_secret_totp", restoredMfaSecret)
}
