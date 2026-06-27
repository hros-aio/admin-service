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

func getClosedAddrForSSO(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	l.Close()
	return addr
}

func TestRedisSSOStateCache_StoreState(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisSSOStateCache(client, logger)

	ctx := context.Background()

	t.Run("successfully stores state with TTL", func(t *testing.T) {
		state := "oauth_state_123"
		nonce := "oidc_nonce_456"
		ttl := 10 * time.Minute

		err := cache.StoreState(ctx, state, nonce, ttl)
		require.NoError(t, err)

		key := "auth:sso_state:" + state
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, nonce, val)

		redisTTL := mr.TTL(key)
		assert.True(t, redisTTL > 9*time.Minute && redisTTL <= 10*time.Minute)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForSSO(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisSSOStateCache(badClient, logger)

		err := badCache.StoreState(ctx, "state", "nonce", 10*time.Minute)
		assert.Error(t, err)
	})
}

func TestRedisSSOStateCache_VerifyAndConsumeState(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisSSOStateCache(client, logger)

	ctx := context.Background()

	t.Run("successfully consumes valid state atomically", func(t *testing.T) {
		state := "oauth_state_123"
		nonce := "oidc_nonce_456"
		key := "auth:sso_state:" + state
		mr.Set(key, nonce)
		mr.SetTTL(key, 10*time.Minute)

		retrieved, err := cache.VerifyAndConsumeState(ctx, state)
		require.NoError(t, err)
		assert.Equal(t, nonce, retrieved)

		// Assert key was deleted in Redis
		assert.False(t, mr.Exists(key))
	})

	t.Run("fails when state is not found or expired", func(t *testing.T) {
		state := "expired_state_999"

		_, err := cache.VerifyAndConsumeState(ctx, state)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidSSOState)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForSSO(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisSSOStateCache(badClient, logger)

		_, err := badCache.VerifyAndConsumeState(ctx, "state")
		assert.Error(t, err)
	})
}
