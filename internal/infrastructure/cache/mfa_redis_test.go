package cache

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisMFACache_StoreToken(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisMFACache(client, logger)

	ctx := context.Background()

	t.Run("successfully stores token with 5-minute TTL", func(t *testing.T) {
		token := "mfa_sess_abc123"
		adminID := "admin_user_uuid_123"

		err := cache.StoreToken(ctx, token, adminID)
		require.NoError(t, err)

		key := "auth:mfa_token:" + token
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, adminID, val)

		redisTTL := mr.TTL(key)
		assert.True(t, redisTTL > 4*time.Minute && redisTTL <= 5*time.Minute)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddr(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisMFACache(badClient, logger)

		err := badCache.StoreToken(ctx, "token", "id")
		assert.Error(t, err)
	})
}

func TestRedisMFACache_GetAdminID(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisMFACache(client, logger)

	ctx := context.Background()

	t.Run("successfully retrieves admin ID for valid token", func(t *testing.T) {
		token := "mfa_sess_abc123"
		adminID := "admin_user_uuid_123"
		mr.Set("auth:mfa_token:"+token, adminID)

		retrievedID, err := cache.GetAdminID(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, adminID, retrievedID)
	})

	t.Run("returns ErrMFATokenExpired for non-existent or expired token", func(t *testing.T) {
		token := "expired_or_missing_token"

		_, err := cache.GetAdminID(ctx, token)
		assert.ErrorIs(t, err, domainErrors.ErrMFATokenExpired)
	})

	t.Run("returns error on redis connection failure", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddr(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisMFACache(badClient, logger)

		_, err := badCache.GetAdminID(ctx, "token")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, domainErrors.ErrMFATokenExpired)
	})
}

func TestRedisMFACache_DeleteToken(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisMFACache(client, logger)

	ctx := context.Background()

	t.Run("successfully deletes token from Redis", func(t *testing.T) {
		token := "mfa_sess_abc123"
		adminID := "admin_user_uuid_123"
		mr.Set("auth:mfa_token:"+token, adminID)

		err := cache.DeleteToken(ctx, token)
		require.NoError(t, err)

		exists := mr.Exists("auth:mfa_token:" + token)
		assert.False(t, exists)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddr(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisMFACache(badClient, logger)

		err := badCache.DeleteToken(ctx, "token")
		assert.Error(t, err)
	})
}

func getClosedAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	err = l.Close()
	require.NoError(t, err)
	return addr
}
