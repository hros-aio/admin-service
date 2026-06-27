package auth

import (
	"context"
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

func setupInviteTokenTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestGormInviteTokenRepository_Save(t *testing.T) {
	gormDB, mock := setupInviteTokenTestDB(t)
	repo := NewGormInviteTokenRepository(gormDB)

	now := time.Now()
	token := &domain.InviteToken{
		ID:        "token-uuid",
		AdminID:   "admin-uuid",
		Token:     "invite-token-string",
		ExpiresAt: now.Add(48 * time.Hour),
		UsedAt:    nil,
		CreatedBy: "inviter-uuid",
		CreatedAt: now,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invite_tokens" ("id","admin_id","token","inviter_id","expires_at","created_at") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "consumed_at"`)).
			WithArgs(token.ID, token.AdminID, token.Token, token.CreatedBy, token.ExpiresAt, token.CreatedAt).
			WillReturnRows(sqlmock.NewRows([]string{"consumed_at"}).AddRow(nil))
		mock.ExpectCommit()

		err := repo.Save(context.Background(), token)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormInviteTokenRepository_FindByToken(t *testing.T) {
	gormDB, mock := setupInviteTokenTestDB(t)
	repo := NewGormInviteTokenRepository(gormDB)

	tokenStr := "invite-token-string"
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "admin_id", "token", "inviter_id", "expires_at", "consumed_at", "created_at",
		}).AddRow(
			"token-uuid", "admin-uuid", tokenStr, "inviter-uuid", now.Add(48*time.Hour), nil, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invite_tokens" WHERE token = $1 ORDER BY "invite_tokens"."id" LIMIT $2`)).
			WithArgs(tokenStr, 1).
			WillReturnRows(rows)

		res, err := repo.FindByToken(context.Background(), tokenStr)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "token-uuid", res.ID)
		assert.Equal(t, tokenStr, res.Token)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invite_tokens" WHERE token = $1 ORDER BY "invite_tokens"."id" LIMIT $2`)).
			WithArgs(tokenStr, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		res, err := repo.FindByToken(context.Background(), tokenStr)
		assert.ErrorIs(t, err, domainErrors.ErrTokenNotFound)
		assert.Nil(t, res)
	})
}

func TestGormInviteTokenRepository_Update(t *testing.T) {
	gormDB, mock := setupInviteTokenTestDB(t)
	repo := NewGormInviteTokenRepository(gormDB)

	now := time.Now()
	token := &domain.InviteToken{
		ID:        "token-uuid",
		AdminID:   "admin-uuid",
		Token:     "invite-token-string",
		ExpiresAt: now.Add(48 * time.Hour),
		UsedAt:    &now,
		CreatedBy: "inviter-uuid",
		CreatedAt: now,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "invite_tokens" SET "admin_id"=$1,"token"=$2,"inviter_id"=$3,"expires_at"=$4,"consumed_at"=$5,"created_at"=$6 WHERE "id" = $7`)).
			WithArgs(token.AdminID, token.Token, token.CreatedBy, token.ExpiresAt, token.UsedAt, token.CreatedAt, token.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), token)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormInviteTokenRepository_Consume(t *testing.T) {
	tokenStr := "invite-token-string"

	t.Run("success", func(t *testing.T) {
		gormDB, mock := setupInviteTokenTestDB(t)
		repo := NewGormInviteTokenRepository(gormDB)

		now := time.Now()
		expiresAt := now.Add(48 * time.Hour)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invite_tokens" WHERE token = $1 ORDER BY "invite_tokens"."id" LIMIT $2 FOR UPDATE`)).
			WithArgs(tokenStr, 1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "admin_id", "token", "inviter_id", "expires_at", "consumed_at", "created_at",
			}).AddRow("token-uuid", "admin-uuid", tokenStr, "inviter-uuid", expiresAt, nil, now))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "invite_tokens" SET "admin_id"=$1,"token"=$2,"inviter_id"=$3,"expires_at"=$4,"consumed_at"=$5,"created_at"=$6 WHERE "id" = $7`)).
			WithArgs("admin-uuid", tokenStr, "inviter-uuid", expiresAt, sqlmock.AnyArg(), now, "token-uuid").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		res, err := repo.Consume(context.Background(), tokenStr)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, res.IsUsed())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("already used", func(t *testing.T) {
		gormDB, mock := setupInviteTokenTestDB(t)
		repo := NewGormInviteTokenRepository(gormDB)

		now := time.Now()
		consumedAt := now.Add(-5 * time.Minute)
		expiresAt := now.Add(48 * time.Hour)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invite_tokens" WHERE token = $1 ORDER BY "invite_tokens"."id" LIMIT $2 FOR UPDATE`)).
			WithArgs(tokenStr, 1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "admin_id", "token", "inviter_id", "expires_at", "consumed_at", "created_at",
			}).AddRow("token-uuid", "admin-uuid", tokenStr, "inviter-uuid", expiresAt, &consumedAt, now))
		mock.ExpectRollback()

		res, err := repo.Consume(context.Background(), tokenStr)
		assert.ErrorIs(t, err, domainErrors.ErrInviteUsed)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("expired", func(t *testing.T) {
		gormDB, mock := setupInviteTokenTestDB(t)
		repo := NewGormInviteTokenRepository(gormDB)

		now := time.Now()
		expiresAt := now.Add(-5 * time.Minute)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invite_tokens" WHERE token = $1 ORDER BY "invite_tokens"."id" LIMIT $2 FOR UPDATE`)).
			WithArgs(tokenStr, 1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "admin_id", "token", "inviter_id", "expires_at", "consumed_at", "created_at",
			}).AddRow("token-uuid", "admin-uuid", tokenStr, "inviter-uuid", expiresAt, nil, now.Add(-24*time.Hour)))
		mock.ExpectRollback()

		res, err := repo.Consume(context.Background(), tokenStr)
		assert.ErrorIs(t, err, domainErrors.ErrInviteExpired)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		gormDB, mock := setupInviteTokenTestDB(t)
		repo := NewGormInviteTokenRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invite_tokens" WHERE token = $1 ORDER BY "invite_tokens"."id" LIMIT $2 FOR UPDATE`)).
			WithArgs(tokenStr, 1).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectRollback()

		res, err := repo.Consume(context.Background(), tokenStr)
		assert.ErrorIs(t, err, domainErrors.ErrTokenNotFound)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
