// Package cache implements the cache infrastructure layer using Redis.
package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/redis/go-redis/v9"
)

// RedisSSOStateCache implements application/interfaces.SSOStateCache using Redis.
type RedisSSOStateCache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisSSOStateCache creates a new RedisSSOStateCache.
func NewRedisSSOStateCache(client *redis.Client, logger *slog.Logger) interfaces.SSOStateCache {
	return &RedisSSOStateCache{
		client: client,
		logger: logger,
	}
}

const (
	ssoStateKeyPrefix = "auth:sso_state:"

	consumeStateScript = `
local val = redis.call('GET', KEYS[1])
if not val then
    return nil
else
    redis.call('DEL', KEYS[1])
    return val
end
`
)

// StoreState caches the nonce/value associated with the OAuth/OIDC state for a specific TTL.
func (r *RedisSSOStateCache) StoreState(ctx context.Context, state string, nonce string, ttl time.Duration) error {
	key := ssoStateKeyPrefix + state
	err := r.client.Set(ctx, key, nonce, ttl).Err()
	if err != nil {
		r.logger.ErrorContext(
			ctx, "failed to store SSO state in Redis",
			slog.String("event", "sso_state_cache_redis.store_failed"),
			slog.String("key", "auth:sso_state:[REDACTED]"),
			slog.Any("error", err),
		)
		return fmt.Errorf("redis set: %w", err)
	}

	r.logger.InfoContext(
		ctx, "successfully stored SSO state in Redis",
		slog.String("event", "sso_state_cache_redis.store_success"),
		slog.String("key", "auth:sso_state:[REDACTED]"),
		slog.Duration("ttl", ttl),
	)
	return nil
}

// VerifyAndConsumeState atomically retrieves the cached nonce/value associated with the OAuth/OIDC state and deletes it.
func (r *RedisSSOStateCache) VerifyAndConsumeState(ctx context.Context, state string) (string, error) {
	key := ssoStateKeyPrefix + state
	res, err := r.client.Eval(ctx, consumeStateScript, []string{key}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.logger.WarnContext(
				ctx, "SSO state not found or expired in Redis",
				slog.String("event", "sso_state_cache_redis.consume_expired"),
				slog.String("key", "auth:sso_state:[REDACTED]"),
			)
			return "", domainErrors.ErrInvalidSSOState
		}
		r.logger.ErrorContext(
			ctx, "failed to consume SSO state from Redis",
			slog.String("event", "sso_state_cache_redis.consume_failed"),
			slog.String("key", "auth:sso_state:[REDACTED]"),
			slog.Any("error", err),
		)
		return "", fmt.Errorf("redis eval: %w", err)
	}

	nonce, ok := res.(string)
	if !ok {
		if res == nil {
			r.logger.WarnContext(
				ctx, "SSO state not found or expired in Redis (nil result)",
				slog.String("event", "sso_state_cache_redis.consume_expired"),
				slog.String("key", "auth:sso_state:[REDACTED]"),
			)
			return "", domainErrors.ErrInvalidSSOState
		}
		return "", fmt.Errorf("unexpected Redis response type: %T", res)
	}

	r.logger.InfoContext(
		ctx, "successfully consumed SSO state from Redis",
		slog.String("event", "sso_state_cache_redis.consume_success"),
		slog.String("key", "auth:sso_state:[REDACTED]"),
	)
	return nonce, nil
}
