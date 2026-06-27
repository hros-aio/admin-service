package auth

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestGormAdminUserRepository_FindByEmail(t *testing.T) {
	gormDB, mock := setupTestDB(t)
	repo := NewGormAdminUserRepository(gormDB)

	email := "admin@example.com"
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role_id", "status",
			"mfa_enabled", "mfa_secret", "last_login_at", "fail_count",
			"locked_until", "invited_by", "created_at", "updated_at",
		}).AddRow(
			"user-uuid", "Admin User", email, "hashed-password", "role-uuid", "active",
			true, "secret", &now, 0, nil, nil, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE email = $1 ORDER BY "admin_users"."id" LIMIT $2`)).
			WithArgs(email, 1).
			WillReturnRows(rows)

		user, err := repo.FindByEmail(context.Background(), email)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, domain.AdminUserStatusActive, user.Status)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE email = $1 ORDER BY "admin_users"."id" LIMIT $2`)).
			WithArgs(email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.FindByEmail(context.Background(), email)

		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE email = $1 ORDER BY "admin_users"."id" LIMIT $2`)).
			WithArgs(email, 1).
			WillReturnError(sql.ErrConnDone)

		user, err := repo.FindByEmail(context.Background(), email)

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestGormAdminUserRepository_UpdatePassword(t *testing.T) {
	adminID := "admin-uuid"
	newHash := "new-hashed-password"

	t.Run("success", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewGormAdminUserRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "admin_users" SET "password_hash"=$1,"updated_at"=$2 WHERE id = $3`)).
			WithArgs(newHash, sqlmock.AnyArg(), adminID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdatePassword(context.Background(), adminID, newHash)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewGormAdminUserRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "admin_users" SET "password_hash"=$1,"updated_at"=$2 WHERE id = $3`)).
			WithArgs(newHash, sqlmock.AnyArg(), adminID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdatePassword(context.Background(), adminID, newHash)
		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewGormAdminUserRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "admin_users" SET "password_hash"=$1,"updated_at"=$2 WHERE id = $3`)).
			WithArgs(newHash, sqlmock.AnyArg(), adminID).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.UpdatePassword(context.Background(), adminID, newHash)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormAdminUserRepository_ActivateAccount(t *testing.T) {
	adminID := "admin-uuid"
	newHash := "new-hashed-password"

	t.Run("success", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewGormAdminUserRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "admin_users" SET "password_hash"=$1,"status"=$2,"updated_at"=$3 WHERE id = $4 AND status = $5`)).
			WithArgs(newHash, "active", sqlmock.AnyArg(), adminID, "pending").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.ActivateAccount(context.Background(), adminID, newHash)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found or not pending", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewGormAdminUserRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "admin_users" SET "password_hash"=$1,"status"=$2,"updated_at"=$3 WHERE id = $4 AND status = $5`)).
			WithArgs(newHash, "active", sqlmock.AnyArg(), adminID, "pending").
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.ActivateAccount(context.Background(), adminID, newHash)
		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		repo := NewGormAdminUserRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "admin_users" SET "password_hash"=$1,"status"=$2,"updated_at"=$3 WHERE id = $4 AND status = $5`)).
			WithArgs(newHash, "active", sqlmock.AnyArg(), adminID, "pending").
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.ActivateAccount(context.Background(), adminID, newHash)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormAdminUserRepository_FindByEmailOrSSO(t *testing.T) {
	gormDB, mock := setupTestDB(t)
	repo := NewGormAdminUserRepository(gormDB)

	email := "admin@example.com"
	ssoProvider := "google"
	ssoID := "sso-id-123"
	now := time.Now()

	t.Run("success match both to same user", func(t *testing.T) {
		rows1 := sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role_id", "status",
			"mfa_enabled", "mfa_secret", "last_login_at", "fail_count",
			"locked_until", "invited_by", "sso_identity_id", "sso_provider", "created_at", "updated_at",
		}).AddRow(
			"user-uuid", "Admin User", email, "hashed-password", "role-uuid", "active",
			true, "secret", &now, 0, nil, nil, &ssoID, &ssoProvider, now, now,
		)

		rows2 := sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role_id", "status",
			"mfa_enabled", "mfa_secret", "last_login_at", "fail_count",
			"locked_until", "invited_by", "sso_identity_id", "sso_provider", "created_at", "updated_at",
		}).AddRow(
			"user-uuid", "Admin User", email, "hashed-password", "role-uuid", "active",
			true, "secret", &now, 0, nil, nil, &ssoID, &ssoProvider, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE email = $1 ORDER BY "admin_users"."id" LIMIT $2`)).
			WithArgs(email, 1).
			WillReturnRows(rows1)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE sso_provider = $1 AND sso_identity_id = $2 ORDER BY "admin_users"."id" LIMIT $3`)).
			WithArgs(ssoProvider, ssoID, 1).
			WillReturnRows(rows2)

		user, err := repo.FindByEmailOrSSO(context.Background(), email, ssoProvider, ssoID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, ssoID, user.SSOIdentityID)
		assert.Equal(t, ssoProvider, user.SSOProvider)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success match email only", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role_id", "status",
			"mfa_enabled", "mfa_secret", "last_login_at", "fail_count",
			"locked_until", "invited_by", "sso_identity_id", "sso_provider", "created_at", "updated_at",
		}).AddRow(
			"user-uuid", "Admin User", email, "hashed-password", "role-uuid", "active",
			true, "secret", &now, 0, nil, nil, nil, nil, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE email = $1 ORDER BY "admin_users"."id" LIMIT $2`)).
			WithArgs(email, 1).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE sso_provider = $1 AND sso_identity_id = $2 ORDER BY "admin_users"."id" LIMIT $3`)).
			WithArgs(ssoProvider, ssoID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.FindByEmailOrSSO(context.Background(), email, ssoProvider, ssoID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Empty(t, user.SSOIdentityID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("conflict match - email and SSO resolve to different users", func(t *testing.T) {
		rows1 := sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role_id", "status",
			"mfa_enabled", "mfa_secret", "last_login_at", "fail_count",
			"locked_until", "invited_by", "sso_identity_id", "sso_provider", "created_at", "updated_at",
		}).AddRow(
			"user-uuid-1", "Admin User 1", email, "hashed-password", "role-uuid", "active",
			true, "secret", &now, 0, nil, nil, nil, nil, now, now,
		)

		rows2 := sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role_id", "status",
			"mfa_enabled", "mfa_secret", "last_login_at", "fail_count",
			"locked_until", "invited_by", "sso_identity_id", "sso_provider", "created_at", "updated_at",
		}).AddRow(
			"user-uuid-2", "Admin User 2", "other@example.com", "hashed-password", "role-uuid", "active",
			true, "secret", &now, 0, nil, nil, &ssoID, &ssoProvider, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE email = $1 ORDER BY "admin_users"."id" LIMIT $2`)).
			WithArgs(email, 1).
			WillReturnRows(rows1)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE sso_provider = $1 AND sso_identity_id = $2 ORDER BY "admin_users"."id" LIMIT $3`)).
			WithArgs(ssoProvider, ssoID, 1).
			WillReturnRows(rows2)

		user, err := repo.FindByEmailOrSSO(context.Background(), email, ssoProvider, ssoID)

		assert.Nil(t, user)
		assert.ErrorIs(t, err, domainErrors.ErrIdentityConflict)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE email = $1 ORDER BY "admin_users"."id" LIMIT $2`)).
			WithArgs(email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "admin_users" WHERE sso_provider = $1 AND sso_identity_id = $2 ORDER BY "admin_users"."id" LIMIT $3`)).
			WithArgs(ssoProvider, ssoID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.FindByEmailOrSSO(context.Background(), email, ssoProvider, ssoID)

		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

