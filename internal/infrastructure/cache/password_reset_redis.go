// Package cache implements the cache infrastructure layer using Redis.
package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/redis/go-redis/v9"
)

// RedisPasswordResetCache implements application/interfaces.PasswordResetCache using Redis.
type RedisPasswordResetCache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisPasswordResetCache creates a new RedisPasswordResetCache.
func NewRedisPasswordResetCache(client *redis.Client, logger *slog.Logger) interfaces.PasswordResetCache {
	return &RedisPasswordResetCache{
		client: client,
		logger: logger,
	}
}

const (
	passwordResetKeyPrefix = "auth:reset_token:"

	consumeTokenScript = `
local val = redis.call('GET', KEYS[1])
if not val then
    return 'expired'
elseif val == 'used' then
    return 'used'
else
    redis.call('SET', KEYS[1], 'used', 'EX', 300)
    return val
end
`
)

// StoreToken associates a reset token with an admin's ID for a specific TTL.
// It enforces a 60-minute TTL or uses the passed duration if specified.
func (r *RedisPasswordResetCache) StoreToken(ctx context.Context, token string, adminID string, ttl time.Duration) error {
	finalTTL := ttl
	if finalTTL == 0 {
		finalTTL = 60 * time.Minute
	}
	key := passwordResetKeyPrefix + token
	err := r.client.Set(ctx, key, adminID, finalTTL).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to store password reset token in Redis",
			slog.String("event", "password_reset_cache_redis.store_failed"),
			slog.String("key", "auth:reset_token:[REDACTED]"),
			slog.Any("error", err),
		)
		return fmt.Errorf("redis set: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully stored password reset token in Redis",
		slog.String("event", "password_reset_cache_redis.store_success"),
		slog.String("key", "auth:reset_token:[REDACTED]"),
		slog.Duration("ttl", finalTTL),
	)
	return nil
}

// ConsumeToken atomically retrieves the cached admin ID associated with the reset token and marks it as used.
func (r *RedisPasswordResetCache) ConsumeToken(ctx context.Context, token string) (string, error) {
	key := passwordResetKeyPrefix + token
	res, err := r.client.Eval(ctx, consumeTokenScript, []string{key}).Result()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to consume password reset token from Redis",
			slog.String("event", "password_reset_cache_redis.consume_failed"),
			slog.String("key", "auth:reset_token:[REDACTED]"),
			slog.Any("error", err),
		)
		return "", fmt.Errorf("redis eval: %w", err)
	}

	status, ok := res.(string)
	if !ok {
		return "", fmt.Errorf("unexpected Redis response type: %T", res)
	}

	switch status {
	case "expired":
		r.logger.WarnContext(ctx, "password reset token not found or expired in Redis",
			slog.String("event", "password_reset_cache_redis.consume_expired"),
			slog.String("key", "auth:reset_token:[REDACTED]"),
		)
		return "", domainErrors.ErrTokenExpired
	case "used":
		r.logger.WarnContext(ctx, "password reset token already consumed in Redis",
			slog.String("event", "password_reset_cache_redis.consume_used"),
			slog.String("key", "auth:reset_token:[REDACTED]"),
		)
		return "", domainErrors.ErrTokenUsed
	default:
		r.logger.InfoContext(ctx, "successfully consumed password reset token from Redis",
			slog.String("event", "password_reset_cache_redis.consume_success"),
			slog.String("key", "auth:reset_token:[REDACTED]"),
		)
		return status, nil
	}
}
