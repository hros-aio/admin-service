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
