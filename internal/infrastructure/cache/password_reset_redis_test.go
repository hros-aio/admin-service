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

func TestRedisPasswordResetCache_ConsumeToken(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisPasswordResetCache(client, logger)

	ctx := context.Background()

	t.Run("successfully consumes valid token and sets state to used with 5m TTL", func(t *testing.T) {
		token := "reset_token_123"
		adminID := "admin_user_uuid_123"
		key := "auth:reset_token:" + token
		mr.Set(key, adminID)
		mr.SetTTL(key, 60*time.Minute)

		retrievedID, err := cache.ConsumeToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, adminID, retrievedID)

		// Assert value is now "used"
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, "used", val)

		// Assert TTL is 5m (300s)
		ttl := mr.TTL(key)
		assert.True(t, ttl > 4*time.Minute && ttl <= 5*time.Minute)
	})

	t.Run("returns ErrTokenUsed when token is already consumed", func(t *testing.T) {
		token := "reset_token_already_used"
		key := "auth:reset_token:" + token
		mr.Set(key, "used")

		_, err := cache.ConsumeToken(ctx, token)
		assert.ErrorIs(t, err, domainErrors.ErrTokenUsed)
	})

	t.Run("returns ErrTokenExpired for non-existent or expired token", func(t *testing.T) {
		token := "expired_or_missing_token"

		_, err := cache.ConsumeToken(ctx, token)
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

		_, err := badCache.ConsumeToken(ctx, "token")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, domainErrors.ErrTokenExpired)
		assert.NotErrorIs(t, err, domainErrors.ErrTokenUsed)
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
