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

// RedisMFACache implements application/interfaces.MFACache using Redis.
type RedisMFACache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisMFACache creates a new RedisMFACache.
func NewRedisMFACache(client *redis.Client, logger *slog.Logger) interfaces.MFACache {
	return &RedisMFACache{
		client: client,
		logger: logger,
	}
}

const (
	mfaKeyPrefix = "auth:mfa_token:"
)

// StoreToken caches the admin ID associated with the MFA token.
func (r *RedisMFACache) StoreToken(ctx context.Context, mfaToken string, adminID string) error {
	const tokenTTL = 5 * time.Minute
	key := mfaKeyPrefix + mfaToken
	err := r.client.Set(ctx, key, adminID, tokenTTL).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to store MFA token in Redis",
			slog.String("event", "mfa_cache_redis.store_failed"),
			slog.String("key", "auth:mfa_token:[REDACTED]"),
			slog.Any("error", err),
		)
		return fmt.Errorf("redis set: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully stored MFA token in Redis",
		slog.String("event", "mfa_cache_redis.store_success"),
		slog.String("key", "auth:mfa_token:[REDACTED]"),
		slog.Duration("ttl", tokenTTL),
	)
	return nil
}

// GetAdminID retrieves the cached admin ID associated with the MFA token.
// It returns domainErrors.ErrMFATokenExpired if the token is not found or has expired.
func (r *RedisMFACache) GetAdminID(ctx context.Context, mfaToken string) (string, error) {
	key := mfaKeyPrefix + mfaToken
	adminID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.logger.WarnContext(ctx, "MFA token not found or expired in Redis",
				slog.String("event", "mfa_cache_redis.get_expired"),
				slog.String("key", "auth:mfa_token:[REDACTED]"),
			)
			return "", domainErrors.ErrMFATokenExpired
		}
		r.logger.ErrorContext(ctx, "failed to retrieve MFA token from Redis",
			slog.String("event", "mfa_cache_redis.get_failed"),
			slog.String("key", "auth:mfa_token:[REDACTED]"),
			slog.Any("error", err),
		)
		return "", fmt.Errorf("redis get: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully retrieved MFA token from Redis",
		slog.String("event", "mfa_cache_redis.get_success"),
		slog.String("key", "auth:mfa_token:[REDACTED]"),
	)
	return adminID, nil
}

// DeleteToken invalidates/removes the cached admin ID for the given MFA token.
func (r *RedisMFACache) DeleteToken(ctx context.Context, mfaToken string) error {
	key := mfaKeyPrefix + mfaToken
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to delete MFA token from Redis",
			slog.String("event", "mfa_cache_redis.delete_failed"),
			slog.String("key", "auth:mfa_token:[REDACTED]"),
			slog.Any("error", err),
		)
		return fmt.Errorf("redis del: %w", err)
	}

	r.logger.InfoContext(ctx, "successfully deleted MFA token from Redis",
		slog.String("event", "mfa_cache_redis.delete_success"),
		slog.String("key", "auth:mfa_token:[REDACTED]"),
	)
	return nil
}
