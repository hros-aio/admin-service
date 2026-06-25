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

// GetAdminID retrieves the cached admin ID associated with the reset token.
// It returns domainErrors.ErrTokenExpired if the token is not found or has expired.
func (r *RedisPasswordResetCache) GetAdminID(ctx context.Context, token string) (string, error) {
	key := passwordResetKeyPrefix + token
	adminID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.logger.WarnContext(ctx, "password reset token not found or expired in Redis",
				slog.String("event", "password_reset_cache_redis.get_expired"),
				slog.String("key", "auth:reset_token:[REDACTED]"),
			)
			return "", domainErrors.ErrTokenExpired
		}
		r.logger.ErrorContext(ctx, "failed to retrieve password reset token from Redis",
			slog.String("event", "password_reset_cache_redis.get_failed"),
			slog.String("key", "auth:reset_token:[REDACTED]"),
			slog.Any("error", err),
		)
		return "", fmt.Errorf("redis get: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully retrieved password reset token from Redis",
		slog.String("event", "password_reset_cache_redis.get_success"),
		slog.String("key", "auth:reset_token:[REDACTED]"),
	)
	return adminID, nil
}

// DeleteToken invalidates/removes the cached reset token.
func (r *RedisPasswordResetCache) DeleteToken(ctx context.Context, token string) error {
	key := passwordResetKeyPrefix + token
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to delete password reset token from Redis",
			slog.String("event", "password_reset_cache_redis.delete_failed"),
			slog.String("key", "auth:reset_token:[REDACTED]"),
			slog.Any("error", err),
		)
		return fmt.Errorf("redis del: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully deleted password reset token from Redis",
		slog.String("event", "password_reset_cache_redis.delete_success"),
		slog.String("key", "auth:reset_token:[REDACTED]"),
	)
	return nil
}
