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

func getClosedAddrForWebAuthn(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	err = l.Close()
	require.NoError(t, err)
	return addr
}

func TestRedisWebAuthnChallengeCache_StoreChallenge(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisWebAuthnChallengeCache(client, logger)

	ctx := context.Background()

	t.Run("successfully stores challenge with TTL", func(t *testing.T) {
		challenge := []byte("secure_challenge_data_xyz")
		ttl := 5 * time.Minute

		err := cache.StoreChallenge(ctx, testEmail, challenge, ttl)
		require.NoError(t, err)

		key := "auth:webauthn_challenge:" + testEmail
		val, err := mr.Get(key)
		require.NoError(t, err)
		assert.Equal(t, string(challenge), val)

		redisTTL := mr.TTL(key)
		assert.True(t, redisTTL > 4*time.Minute && redisTTL <= 5*time.Minute)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForWebAuthn(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisWebAuthnChallengeCache(badClient, logger)

		err := badCache.StoreChallenge(ctx, testEmail, []byte("challenge"), 5*time.Minute)
		assert.Error(t, err)
	})
}

func TestRedisWebAuthnChallengeCache_GetChallenge(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisWebAuthnChallengeCache(client, logger)

	ctx := context.Background()

	t.Run("successfully retrieves valid challenge", func(t *testing.T) {
		challenge := []byte("secure_challenge_data_xyz")
		key := "auth:webauthn_challenge:" + testEmail
		err := mr.Set(key, string(challenge))
		require.NoError(t, err)

		retrieved, err := cache.GetChallenge(ctx, testEmail)
		require.NoError(t, err)
		assert.Equal(t, challenge, retrieved)
	})

	t.Run("fails when challenge is not found or expired", func(t *testing.T) {
		email := "missing@hros.com"

		_, err := cache.GetChallenge(ctx, email)
		assert.ErrorIs(t, err, domainErrors.ErrChallengeNotFoundOrExpired)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForWebAuthn(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisWebAuthnChallengeCache(badClient, logger)

		_, err := badCache.GetChallenge(ctx, testEmail)
		assert.Error(t, err)
	})
}

func TestRedisWebAuthnChallengeCache_DeleteChallenge(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisWebAuthnChallengeCache(client, logger)

	ctx := context.Background()

	t.Run("successfully deletes challenge", func(t *testing.T) {
		challenge := []byte("secure_challenge_data_xyz")
		key := "auth:webauthn_challenge:" + testEmail
		err := mr.Set(key, string(challenge))
		require.NoError(t, err)

		err = cache.DeleteChallenge(ctx, testEmail)
		require.NoError(t, err)

		// Assert key was deleted
		assert.False(t, mr.Exists(key))
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForWebAuthn(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisWebAuthnChallengeCache(badClient, logger)

		err := badCache.DeleteChallenge(ctx, testEmail)
		assert.Error(t, err)
	})
}

func TestRedisWebAuthnChallengeCache_VerifyAndConsumeChallenge(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := NewRedisWebAuthnChallengeCache(client, logger)

	ctx := context.Background()

	t.Run("successfully consumes challenge atomically", func(t *testing.T) {
		challenge := []byte("secure_challenge_data_xyz")
		key := "auth:webauthn_challenge:" + testEmail
		err := mr.Set(key, string(challenge))
		require.NoError(t, err)
		mr.SetTTL(key, 5*time.Minute)

		retrieved, err := cache.VerifyAndConsumeChallenge(ctx, testEmail)
		require.NoError(t, err)
		assert.Equal(t, challenge, retrieved)

		// Assert key was deleted in Redis after consumption
		assert.False(t, mr.Exists(key))
	})

	t.Run("fails when challenge is not found or expired", func(t *testing.T) {
		email := "expired_challenge@hros.com"

		_, err := cache.VerifyAndConsumeChallenge(ctx, email)
		assert.ErrorIs(t, err, domainErrors.ErrChallengeNotFoundOrExpired)
	})

	t.Run("error when client fails", func(t *testing.T) {
		badClient := redis.NewClient(&redis.Options{
			Addr:        getClosedAddrForWebAuthn(t),
			MaxRetries:  -1,
			DialTimeout: 10 * time.Millisecond,
		})
		defer func() { _ = badClient.Close() }()

		badCache := NewRedisWebAuthnChallengeCache(badClient, logger)

		_, err := badCache.VerifyAndConsumeChallenge(ctx, testEmail)
		assert.Error(t, err)
	})
}
