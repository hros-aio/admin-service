package integration

import (
	"context"
	"os"
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

// runMigrationSQLFile parses and executes SQL migration scripts statement-by-statement,
// correctly ignoring semicolons inside dollar-quoted ($$) blocks.
func runMigrationSQLFile(t *testing.T, db *gorm.DB, filepath string) {
	t.Helper()
	content, err := os.ReadFile(filepath)
	require.NoError(t, err)

	var statements []string
	var currentStatement strings.Builder
	inDollarQuote := false

	runes := []rune(string(content))
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '$' && i+1 < len(runes) && runes[i+1] == '$' {
			inDollarQuote = !inDollarQuote
			currentStatement.WriteRune('$')
			currentStatement.WriteRune('$')
			i++
			continue
		}
		if r == ';' && !inDollarQuote {
			stmt := strings.TrimSpace(currentStatement.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			currentStatement.Reset()
			continue
		}
		currentStatement.WriteRune(r)
	}
	stmt := strings.TrimSpace(currentStatement.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	for _, stmt := range statements {
		lines := strings.Split(stmt, "\n")
		var cleanLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "--") {
				continue
			}
			cleanLines = append(cleanLines, line)
		}
		execStmt := strings.TrimSpace(strings.Join(cleanLines, "\n"))
		if execStmt == "" {
			continue
		}
		err := db.Exec(execStmt).Error
		require.NoError(t, err, "failed to execute statement: %s", execStmt)
	}
}

// setupMigrationTest starts a PostgreSQL container, executes baseline migrations,
// seeds a test user with old-style MFA, and returns database references.
func setupMigrationTest(t *testing.T) (*gorm.DB, string, string, func()) {
	t.Helper()
	ctx := context.Background()

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

	roleID := domain.NewUUID()
	err = db.Exec("INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		roleID, "Super Admin", "Super Admin Role", true).Error
	if err != nil {
		cleanup()
		require.NoError(t, err)
	}

	adminUserID := domain.NewUUID()
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status, mfa_enabled, mfa_secret) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		adminUserID, "MFA Test Admin", "mfa-admin@hros.com", "hash", roleID, "active", true, "my_super_secret_totp").Error
	if err != nil {
		cleanup()
		require.NoError(t, err)
	}

	return db, migDir, adminUserID, cleanup
}

func TestMFAMigrationUpFlow(t *testing.T) {
	db, migDir, adminUserID, cleanup := setupMigrationTest(t)
	defer cleanup()

	// Run up migration the first time
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))

	// Verify that mfa_secret column is dropped (selecting it should error)
	var dummy string
	err := db.Raw("SELECT mfa_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	require.Error(t, err)
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

	// Run up migration a second time to verify idempotency
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))

	// Verify that mfa_secret is still dropped
	err = db.Raw("SELECT mfa_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "mfa_secret") || strings.Contains(err.Error(), "does not exist"))

	// Verify that data in totp_secret remains consistent and is not overwritten or cleared
	var totpSecretSecond string
	err = db.Raw("SELECT totp_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&totpSecretSecond).Error
	require.NoError(t, err)
	assert.Equal(t, "my_super_secret_totp", totpSecretSecond)
}

func TestMFAMigrationDownFlow(t *testing.T) {
	db, migDir, adminUserID, cleanup := setupMigrationTest(t)
	defer cleanup()

	// Apply up migration first to place database in target state
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))

	// Run down migration the first time
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.down.sql"))

	// Verify totp_secret column is dropped (selecting it should error)
	var dummy string
	err := db.Raw("SELECT totp_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "totp_secret") || strings.Contains(err.Error(), "does not exist"), "Expected error about totp_secret column not existing, got: %v", err)

	// Verify webauthn_credentials column is dropped (selecting it should error)
	err = db.Raw("SELECT webauthn_credentials FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "webauthn_credentials") || strings.Contains(err.Error(), "does not exist"), "Expected error about webauthn_credentials column not existing, got: %v", err)

	// Verify that mfa_secret column is restored with correct data
	var restoredMfaSecret string
	err = db.Raw("SELECT mfa_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&restoredMfaSecret).Error
	require.NoError(t, err)
	assert.Equal(t, "my_super_secret_totp", restoredMfaSecret)

	// Run down migration a second time to verify idempotency
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.down.sql"))

	// Verify totp_secret remains dropped
	err = db.Raw("SELECT totp_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "totp_secret") || strings.Contains(err.Error(), "does not exist"))

	// Verify webauthn_credentials remains dropped
	err = db.Raw("SELECT webauthn_credentials FROM admin_users WHERE id = ?", adminUserID).Scan(&dummy).Error
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "webauthn_credentials") || strings.Contains(err.Error(), "does not exist"))

	// Verify restored mfa_secret remains consistent
	var restoredMfaSecretSecond string
	err = db.Raw("SELECT mfa_secret FROM admin_users WHERE id = ?", adminUserID).Scan(&restoredMfaSecretSecond).Error
	require.NoError(t, err)
	assert.Equal(t, "my_super_secret_totp", restoredMfaSecretSecond)
}
