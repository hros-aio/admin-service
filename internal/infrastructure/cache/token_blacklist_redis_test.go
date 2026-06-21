package cache

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisTokenBlacklist_Add(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	// Discard log output during tests
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	blacklist := NewRedisTokenBlacklist(client, logger)

	ctx := context.Background()

	t.Run("successful blacklist add with normal TTL", func(t *testing.T) {
		token := "normal-token-jti"
		ttl := 10 * time.Minute

		err := blacklist.Add(ctx, token, ttl)
		require.NoError(t, err)

		// Assert value and TTL in miniredis
		key := "blacklist:jti:" + token
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, "1", val)

		redisTTL := mr.TTL(key)
		// miniredis TTL is duration. Assert it is close to 10m
		assert.True(t, redisTTL > 9*time.Minute && redisTTL <= 10*time.Minute)
	})

	t.Run("successful blacklist add with capped TTL", func(t *testing.T) {
		token := "long-token-jti"
		ttl := 30 * time.Minute // Longer than maxTTL (15 mins)

		err := blacklist.Add(ctx, token, ttl)
		require.NoError(t, err)

		key := "blacklist:jti:" + token
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, "1", val)

		redisTTL := mr.TTL(key)
		// Assert TTL is capped at maxTTL (15 minutes)
		assert.True(t, redisTTL > 14*time.Minute && redisTTL <= 15*time.Minute)
	})
}

func TestRedisTokenBlacklist_Exists(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	blacklist := NewRedisTokenBlacklist(client, logger)

	ctx := context.Background()

	t.Run("token exists in blacklist", func(t *testing.T) {
		token := "exists-token-jti"
		_ = mr.Set("blacklist:jti:"+token, "1")

		exists, err := blacklist.Exists(ctx, token)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("token does not exist in blacklist", func(t *testing.T) {
		token := "non-existent-token-jti"

		exists, err := blacklist.Exists(ctx, token)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("graceful degradation on Redis connection error", func(t *testing.T) {
		// Use a client pointing to a closed miniredis address with fast fail options
		badClient := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			MaxRetries:  -1, // No retries
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badBlacklist := NewRedisTokenBlacklist(badClient, logger)

		exists, err := badBlacklist.Exists(ctx, "any-token")
		// Graceful degradation: should return false and no error (nil)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
