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

func TestRedisPasswordResetCache_StoreToken(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisPasswordResetCache(client, logger)

	ctx := context.Background()

	t.Run("successfully stores token with 60-minute TTL", func(t *testing.T) {
		token := "reset_token_123"
		adminID := "admin_user_uuid_123"

		err := cache.StoreToken(ctx, token, adminID, 0)
		require.NoError(t, err)

		key := "auth:reset_token:" + token
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, adminID, val)

		redisTTL := mr.TTL(key)
		assert.True(t, redisTTL > 59*time.Minute && redisTTL <= 60*time.Minute)
	})

	t.Run("successfully stores token with custom TTL", func(t *testing.T) {
		token := "reset_token_456"
		adminID := "admin_user_uuid_456"
		customTTL := 10 * time.Minute

		err := cache.StoreToken(ctx, token, adminID, customTTL)
		require.NoError(t, err)

		key := "auth:reset_token:" + token
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, adminID, val)

		redisTTL := mr.TTL(key)
		assert.True(t, redisTTL > 9*time.Minute && redisTTL <= 10*time.Minute)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForReset(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisPasswordResetCache(badClient, logger)

		err := badCache.StoreToken(ctx, "token", "id", 0)
		assert.Error(t, err)
	})
}

func TestRedisPasswordResetCache_GetAdminID(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisPasswordResetCache(client, logger)

	ctx := context.Background()

	t.Run("successfully retrieves admin ID for valid token", func(t *testing.T) {
		token := "reset_token_123"
		adminID := "admin_user_uuid_123"
		mr.Set("auth:reset_token:"+token, adminID)

		retrievedID, err := cache.GetAdminID(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, adminID, retrievedID)
	})

	t.Run("returns ErrTokenExpired for non-existent or expired token", func(t *testing.T) {
		token := "expired_or_missing_token"

		_, err := cache.GetAdminID(ctx, token)
		assert.ErrorIs(t, err, domainErrors.ErrTokenExpired)
	})

	t.Run("returns error on redis connection failure", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForReset(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisPasswordResetCache(badClient, logger)

		_, err := badCache.GetAdminID(ctx, "token")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, domainErrors.ErrTokenExpired)
	})
}

func TestRedisPasswordResetCache_DeleteToken(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisPasswordResetCache(client, logger)

	ctx := context.Background()

	t.Run("successfully deletes token from Redis", func(t *testing.T) {
		token := "reset_token_123"
		adminID := "admin_user_uuid_123"
		mr.Set("auth:reset_token:"+token, adminID)

		err := cache.DeleteToken(ctx, token)
		require.NoError(t, err)

		exists := mr.Exists("auth:reset_token:" + token)
		assert.False(t, exists)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForReset(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisPasswordResetCache(badClient, logger)

		err := badCache.DeleteToken(ctx, "token")
		assert.Error(t, err)
	})
}

func getClosedAddrForReset(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	err = l.Close()
	require.NoError(t, err)
	return addr
}
