// Package cache implements the cache infrastructure layer using Redis.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/redis/go-redis/v9"
)

// RedisWebAuthnChallengeCache implements application/interfaces.WebAuthnChallengeCache using Redis.
type RedisWebAuthnChallengeCache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisWebAuthnChallengeCache creates a new RedisWebAuthnChallengeCache.
func NewRedisWebAuthnChallengeCache(client *redis.Client, logger *slog.Logger) interfaces.WebAuthnChallengeCache {
	return &RedisWebAuthnChallengeCache{
		client: client,
		logger: logger,
	}
}

const (
	webauthnKeyPrefix = "auth:webauthn_challenge:"

	consumeChallengeScript = `
local val = redis.call('GET', KEYS[1])
if not val then
    return nil
else
    redis.call('DEL', KEYS[1])
    return val
end
`
)

// StoreChallenge caches the challenge bytes associated with a ceremony or session-scoped email.
func (r *RedisWebAuthnChallengeCache) StoreChallenge(ctx context.Context, email string, challenge []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return errors.New("TTL must be positive")
	}
	opaqueID := hashEmail(email)
	redisKey := webauthnKeyPrefix + opaqueID
	err := r.client.Set(ctx, redisKey, challenge, ttl).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to store WebAuthn challenge in Redis",
			slog.String("event", "webauthn_challenge_cache_redis.store_failed"),
			slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
			slog.Any("error", err),
		)
		return fmt.Errorf("redis set: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully stored WebAuthn challenge in Redis",
		slog.String("event", "webauthn_challenge_cache_redis.store_success"),
		slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
		slog.Duration("ttl", ttl),
	)
	return nil
}

// GetChallenge retrieves the cached challenge bytes for the given email.
// It returns domainErrors.ErrChallengeNotFoundOrExpired if the challenge does not exist or has expired.
func (r *RedisWebAuthnChallengeCache) GetChallenge(ctx context.Context, email string) ([]byte, error) {
	opaqueID := hashEmail(email)
	redisKey := webauthnKeyPrefix + opaqueID
	val, err := r.client.Get(ctx, redisKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.logger.WarnContext(ctx, "WebAuthn challenge not found or expired in Redis",
				slog.String("event", "webauthn_challenge_cache_redis.get_expired"),
				slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
			)
			return nil, domainErrors.ErrChallengeNotFoundOrExpired
		}
		r.logger.ErrorContext(ctx, "failed to retrieve WebAuthn challenge from Redis",
			slog.String("event", "webauthn_challenge_cache_redis.get_failed"),
			slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("redis get: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully retrieved WebAuthn challenge from Redis",
		slog.String("event", "webauthn_challenge_cache_redis.get_success"),
		slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
	)
	return val, nil
}

// DeleteChallenge removes the cached challenge for the given email.
func (r *RedisWebAuthnChallengeCache) DeleteChallenge(ctx context.Context, email string) error {
	opaqueID := hashEmail(email)
	redisKey := webauthnKeyPrefix + opaqueID
	err := r.client.Del(ctx, redisKey).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to delete WebAuthn challenge from Redis",
			slog.String("event", "webauthn_challenge_cache_redis.delete_failed"),
			slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
			slog.Any("error", err),
		)
		return fmt.Errorf("redis del: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully deleted WebAuthn challenge from Redis",
		slog.String("event", "webauthn_challenge_cache_redis.delete_success"),
		slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
	)
	return nil
}

// VerifyAndConsumeChallenge atomically retrieves the cached challenge bytes and deletes it.
// It returns domainErrors.ErrChallengeNotFoundOrExpired if the challenge is not found or has expired.
func (r *RedisWebAuthnChallengeCache) VerifyAndConsumeChallenge(ctx context.Context, email string) ([]byte, error) {
	opaqueID := hashEmail(email)
	redisKey := webauthnKeyPrefix + opaqueID
	res, err := r.client.Eval(ctx, consumeChallengeScript, []string{redisKey}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.logger.WarnContext(ctx, "WebAuthn challenge not found or expired in Redis during consume",
				slog.String("event", "webauthn_challenge_cache_redis.consume_expired"),
				slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
			)
			return nil, domainErrors.ErrChallengeNotFoundOrExpired
		}
		r.logger.ErrorContext(ctx, "failed to consume WebAuthn challenge from Redis",
			slog.String("event", "webauthn_challenge_cache_redis.consume_failed"),
			slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("redis eval: %w", err)
	}

	valStr, ok := res.(string)
	if !ok {
		if res == nil {
			r.logger.WarnContext(ctx, "WebAuthn challenge not found or expired in Redis during consume (nil result)",
				slog.String("event", "webauthn_challenge_cache_redis.consume_expired"),
				slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
			)
			return nil, domainErrors.ErrChallengeNotFoundOrExpired
		}
		return nil, fmt.Errorf("unexpected Redis response type: %T", res)
	}

	r.logger.InfoContext(ctx, "successfully consumed WebAuthn challenge from Redis",
		slog.String("event", "webauthn_challenge_cache_redis.consume_success"),
		slog.String("key", "auth:webauthn_challenge:[REDACTED]"),
	)
	return []byte(valStr), nil
}

func hashEmail(email string) string {
	h := sha256.New()
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))
}
