// Package cache implements the cache infrastructure layer using Redis.
package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/redis/go-redis/v9"
)

// RedisTokenBlacklist implements application/interfaces.TokenBlacklist using Redis.
type RedisTokenBlacklist struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisTokenBlacklist creates a new RedisTokenBlacklist.
func NewRedisTokenBlacklist(client *redis.Client, logger *slog.Logger) interfaces.TokenBlacklist {
	return &RedisTokenBlacklist{
		client: client,
		logger: logger,
	}
}

const (
	blacklistKeyPrefix = "blacklist:jti:"
	maxTTL             = 15 * time.Minute
)

// Add stores a JTI/token in the Redis blacklist with a TTL, capped at 15 minutes.
func (r *RedisTokenBlacklist) Add(ctx context.Context, token string, ttl time.Duration) error {
	if ttl > maxTTL {
		ttl = maxTTL
	}

	key := blacklistKeyPrefix + token
	err := r.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to add token to Redis blacklist",
			slog.String("event", "token_blacklist_redis.add_failed"),
			slog.String("key", key),
			slog.Duration("ttl", ttl),
			slog.Any("error", err),
		)
		return fmt.Errorf("redis set: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully blacklisted token",
		slog.String("event", "token_blacklist_redis.add_success"),
		slog.String("key", key),
		slog.Duration("ttl", ttl),
	)
	return nil
}

// Exists checks if a JTI/token is in the Redis blacklist.
// If a connection error occurs, it degrades gracefully by logging the error and returning false.
func (r *RedisTokenBlacklist) Exists(ctx context.Context, token string) (bool, error) {
	key := blacklistKeyPrefix + token
	val, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to query Redis blacklist, degrading gracefully (fail-open)",
			slog.String("event", "token_blacklist_redis.exists_failed"),
			slog.String("key", key),
			slog.Any("error", err),
		)
		// Graceful degradation: return false (fail-open) and nil error so auth check doesn't block users when Redis is down.
		return false, nil
	}

	isBlacklisted := val > 0
	if isBlacklisted {
		r.logger.InfoContext(ctx, "token blacklist check: token is blacklisted",
			slog.String("event", "token_blacklist_redis.exists_true"),
			slog.String("key", key),
		)
	}

	return isBlacklisted, nil
}
