// Package auth implements the authentication repository using GORM.
package auth

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hros/admin-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupSessionTokenTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestGormSessionTokenRepository_Save(t *testing.T) {
	gormDB, mock := setupSessionTokenTestDB(t)
	repo := NewGormSessionTokenRepository(gormDB)

	now := time.Now()
	token := &domain.SessionToken{
		ID:           "test-id",
		AdminID:      "admin-id",
		RefreshToken: "refresh-token",
		ExpiresAt:    now.Add(24 * time.Hour),
		IsPersistent: true,
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
		CreatedAt:    now,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "session_tokens" ("id","admin_id","refresh_token","expires_at","is_persistent","ip_address","user_agent","created_at","revoke_reason") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "revoked_at"`)).
			WithArgs(token.ID, token.AdminID, token.RefreshToken, token.ExpiresAt, token.IsPersistent, token.IPAddress, token.UserAgent, token.CreatedAt, nil).
			WillReturnRows(sqlmock.NewRows([]string{"revoked_at"}).AddRow(nil))
		mock.ExpectCommit()

		err := repo.Save(context.Background(), token)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "session_tokens"`)).
			WillReturnError(gorm.ErrInvalidDB)
		mock.ExpectRollback()

		err := repo.Save(context.Background(), token)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("nil token", func(t *testing.T) {
		err := repo.Save(context.Background(), nil)
		assert.ErrorIs(t, err, gorm.ErrInvalidData)
	})
}

func TestGormSessionTokenRepository_DeleteByToken(t *testing.T) {
	gormDB, mock := setupSessionTokenTestDB(t)
	repo := NewGormSessionTokenRepository(gormDB)
	tokenValue := "test-delete-token"

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "session_tokens" WHERE refresh_token = $1`)).
			WithArgs(tokenValue).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.DeleteByToken(context.Background(), tokenValue)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - idempotent", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "session_tokens" WHERE refresh_token = $1`)).
			WithArgs(tokenValue).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.DeleteByToken(context.Background(), tokenValue)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormSessionTokenRepository_Revoke(t *testing.T) {
	gormDB, mock := setupSessionTokenTestDB(t)
	repo := NewGormSessionTokenRepository(gormDB)
	tokenValue := "test-revoke-token"
	reason := "security violation"

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "session_tokens" SET "revoke_reason"=$1,"revoked_at"=NOW() WHERE refresh_token = $2`)).
			WithArgs(reason, tokenValue).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Revoke(context.Background(), tokenValue, reason)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormSessionTokenRepository_FindByToken(t *testing.T) {
	gormDB, mock := setupSessionTokenTestDB(t)
	repo := NewGormSessionTokenRepository(gormDB)
	tokenValue := "test-token"
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM "session_tokens" WHERE refresh_token = \$1.*`).
			WithArgs(tokenValue, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "admin_id", "refresh_token", "expires_at", "is_persistent", "ip_address", "user_agent", "created_at", "revoked_at", "revoke_reason"}).
				AddRow("test-id", "admin-id", tokenValue, now, true, "127.0.0.1", "test-agent", now, nil, nil))

		session, err := repo.FindByToken(context.Background(), tokenValue)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "test-id", session.ID)
		assert.Equal(t, "admin-id", session.AdminID)
		assert.Equal(t, tokenValue, session.RefreshToken)
		assert.True(t, session.IsPersistent)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM "session_tokens" WHERE refresh_token = \$1.*`).
			WithArgs(tokenValue, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		session, err := repo.FindByToken(context.Background(), tokenValue)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Nil(t, session)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormSessionTokenRepository_UpdateToken(t *testing.T) {
	gormDB, mock := setupSessionTokenTestDB(t)
	repo := NewGormSessionTokenRepository(gormDB)

	now := time.Now()
	token := &domain.SessionToken{
		ID:           "test-id",
		AdminID:      "admin-id",
		RefreshToken: "refresh-token",
		ExpiresAt:    now.Add(24 * time.Hour),
		IsPersistent: true,
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
		CreatedAt:    now,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "session_tokens" SET .* WHERE "id" = \$10`).
			WithArgs(token.AdminID, token.RefreshToken, token.ExpiresAt, token.IsPersistent, token.IPAddress, token.UserAgent, token.CreatedAt, nil, nil, token.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateToken(context.Background(), token)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "session_tokens" SET .* WHERE "id" = \$10`).
			WillReturnError(gorm.ErrInvalidDB)
		mock.ExpectRollback()

		err := repo.UpdateToken(context.Background(), token)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("nil token", func(t *testing.T) {
		err := repo.UpdateToken(context.Background(), nil)
		assert.ErrorIs(t, err, gorm.ErrInvalidData)
	})
}
