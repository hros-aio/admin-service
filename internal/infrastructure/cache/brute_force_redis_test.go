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

const testEmail = "user@hros.io"

func TestRedisBruteForceCache_IncrementFailedAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisBruteForceCache(client, logger)
	ctx := context.Background()
	email := testEmail

	t.Run("first failure sets attempts and TTL", func(t *testing.T) {
		mr.FlushAll()

		attempts, err := cache.IncrementFailedAttempts(ctx, email, 15*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, 1, attempts)

		key := "auth:failed_attempts:" + email
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, "1", val)

		ttl := mr.TTL(key)
		assert.True(t, ttl > 14*time.Minute && ttl <= 15*time.Minute)
	})

	t.Run("subsequent failure increments without resetting TTL", func(t *testing.T) {
		mr.FlushAll()

		// Set initial value with 10 minutes TTL
		key := "auth:failed_attempts:" + email
		err := mr.Set(key, "1")
		require.NoError(t, err)
		mr.SetTTL(key, 10*time.Minute)

		attempts, err := cache.IncrementFailedAttempts(ctx, email, 15*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, 2, attempts)

		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, "2", val)

		// TTL should still be around 10 minutes (not overwritten to 15)
		ttl := mr.TTL(key)
		assert.True(t, ttl > 9*time.Minute && ttl <= 10*time.Minute)
	})

	t.Run("graceful degradation on error", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        "192.0.2.1:6379",
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisBruteForceCache(badClient, logger)
		attempts, err := badCache.IncrementFailedAttempts(ctx, email, 15*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, 0, attempts)
	})
}

func TestRedisBruteForceCache_GetFailedAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisBruteForceCache(client, logger)
	ctx := context.Background()
	email := testEmail
	key := "auth:failed_attempts:" + email

	t.Run("attempts key exists", func(t *testing.T) {
		mr.FlushAll()
		err := mr.Set(key, "3")
		require.NoError(t, err)

		attempts, err := cache.GetFailedAttempts(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("attempts key does not exist", func(t *testing.T) {
		mr.FlushAll()

		attempts, err := cache.GetFailedAttempts(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, 0, attempts)
	})

	t.Run("graceful degradation on error", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        "192.0.2.1:6379",
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisBruteForceCache(badClient, logger)
		attempts, err := badCache.GetFailedAttempts(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, 0, attempts)
	})
}

func TestRedisBruteForceCache_SetLockout(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisBruteForceCache(client, logger)
	ctx := context.Background()
	email := testEmail
	key := "auth:lockout:" + email

	t.Run("successful lockout sets key and TTL", func(t *testing.T) {
		mr.FlushAll()

		err := cache.SetLockout(ctx, email, 30*time.Minute)
		require.NoError(t, err)

		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.NotEmpty(t, val)

		// Assert value parsing
		parsed, err := time.Parse(time.RFC3339, val)
		require.NoError(t, err)
		assert.True(t, parsed.After(time.Now()))

		ttl := mr.TTL(key)
		assert.True(t, ttl > 29*time.Minute && ttl <= 30*time.Minute)
	})

	t.Run("graceful degradation on error", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        "192.0.2.1:6379",
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisBruteForceCache(badClient, logger)
		err := badCache.SetLockout(ctx, email, 30*time.Minute)
		require.NoError(t, err)
	})
}

func TestRedisBruteForceCache_IsLocked(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisBruteForceCache(client, logger)
	ctx := context.Background()
	email := testEmail
	key := "auth:lockout:" + email

	t.Run("is locked with valid time representation", func(t *testing.T) {
		mr.FlushAll()
		future := time.Now().Add(30 * time.Minute).Truncate(time.Second)
		err := mr.Set(key, future.Format(time.RFC3339))
		require.NoError(t, err)

		locked, expiry, err := cache.IsLocked(ctx, email)
		require.NoError(t, err)
		assert.True(t, locked)
		assert.True(t, expiry.Equal(future) || expiry.After(future.Add(-2*time.Second)))
	})

	t.Run("is locked with malformed value - falls back to TTL", func(t *testing.T) {
		mr.FlushAll()
		err := mr.Set(key, "invalid-time")
		require.NoError(t, err)
		mr.SetTTL(key, 25*time.Minute)

		locked, expiry, err := cache.IsLocked(ctx, email)
		require.NoError(t, err)
		assert.True(t, locked)
		assert.True(t, expiry.After(time.Now().Add(24*time.Minute)))
	})

	t.Run("is locked with malformed value and no TTL - fails open", func(t *testing.T) {
		mr.FlushAll()
		err := mr.Set(key, "invalid-time")
		require.NoError(t, err)

		locked, _, err := cache.IsLocked(ctx, email)
		require.NoError(t, err)
		assert.False(t, locked)
	})

	t.Run("is not locked", func(t *testing.T) {
		mr.FlushAll()

		locked, _, err := cache.IsLocked(ctx, email)
		require.NoError(t, err)
		assert.False(t, locked)
	})

	t.Run("graceful degradation on error", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        "192.0.2.1:6379",
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisBruteForceCache(badClient, logger)
		locked, _, err := badCache.IsLocked(ctx, email)
		require.NoError(t, err)
		assert.False(t, locked)
	})
}

func TestRedisBruteForceCache_Reset(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisBruteForceCache(client, logger)
	ctx := context.Background()
	email := testEmail

	keyAttempts := "auth:failed_attempts:" + email
	keyLockout := "auth:lockout:" + email

	t.Run("successful reset clears all keys", func(t *testing.T) {
		mr.FlushAll()
		err1 := mr.Set(keyAttempts, "4")
		require.NoError(t, err1)
		err2 := mr.Set(keyLockout, "locked")
		require.NoError(t, err2)

		err := cache.Reset(ctx, email)
		require.NoError(t, err)

		assert.False(t, mr.Exists(keyAttempts))
		assert.False(t, mr.Exists(keyLockout))
	})

	t.Run("graceful degradation on error", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        "192.0.2.1:6379",
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisBruteForceCache(badClient, logger)
		err := badCache.Reset(ctx, email)
		require.NoError(t, err)
	})
}
