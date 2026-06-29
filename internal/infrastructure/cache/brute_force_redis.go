// Package cache implements the cache infrastructure layer using Redis.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/redis/go-redis/v9"
)

// RedisBruteForceCache implements application/interfaces.BruteForceCache using Redis.
type RedisBruteForceCache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisBruteForceCache creates a new RedisBruteForceCache.
func NewRedisBruteForceCache(client *redis.Client, logger *slog.Logger) interfaces.BruteForceCache {
	return &RedisBruteForceCache{
		client: client,
		logger: logger,
	}
}

const (
	failedAttemptsKeyPrefix = "auth:failed_attempts:"
	lockoutKeyPrefix        = "auth:lockout:"
)

// maskEmail hashes sensitive emails using SHA256 (truncated to 12 chars) to avoid PII leak in logs.
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	h := sha256.New()
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))[:12]
}

// maskKey hashes the email portion of Redis keys to avoid PII leak in logs.
func maskKey(key string) string {
	if len(key) > len(failedAttemptsKeyPrefix) && key[:len(failedAttemptsKeyPrefix)] == failedAttemptsKeyPrefix {
		email := key[len(failedAttemptsKeyPrefix):]
		return failedAttemptsKeyPrefix + maskEmail(email)
	}
	if len(key) > len(lockoutKeyPrefix) && key[:len(lockoutKeyPrefix)] == lockoutKeyPrefix {
		email := key[len(lockoutKeyPrefix):]
		return lockoutKeyPrefix + maskEmail(email)
	}
	return key
}

// IncrementFailedAttempts increments the failed login attempts counter for an email within a sliding window.
// If an error occurs, it logs the failure and degrades gracefully by returning 0 and nil error.
func (r *RedisBruteForceCache) IncrementFailedAttempts(ctx context.Context, email string, window time.Duration) (int, error) {
	key := failedAttemptsKeyPrefix + email
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		r.logger.ErrorContext(
			ctx, "failed to increment failed attempts in Redis, degrading gracefully (fail-open)",
			slog.String("event", "brute_force_redis.increment_failed"),
			slog.String("email", maskEmail(email)),
			slog.Any("error", err),
		)
		return 0, nil
	}

	if val == 1 {
		err = r.client.Expire(ctx, key, window).Err()
		if err != nil {
			r.logger.WarnContext(
				ctx, "failed to set TTL for failed attempts key, deleting key to fail-open",
				slog.String("event", "brute_force_redis.expire_failed"),
				slog.String("key", maskKey(key)),
				slog.Any("error", err),
			)
			_ = r.client.Del(ctx, key).Err()
		}
	}

	r.logger.InfoContext(
		ctx, "successfully incremented failed attempts",
		slog.String("event", "brute_force_redis.increment_success"),
		slog.String("email", maskEmail(email)),
		slog.Int("attempts", int(val)),
	)
	return int(val), nil
}

// GetFailedAttempts returns the current failed login attempts counter for an email.
// If an error occurs, it logs the failure and degrades gracefully by returning 0 and nil error.
func (r *RedisBruteForceCache) GetFailedAttempts(ctx context.Context, email string) (int, error) {
	key := failedAttemptsKeyPrefix + email
	val, err := r.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		r.logger.ErrorContext(
			ctx, "failed to retrieve failed attempts from Redis, degrading gracefully",
			slog.String("event", "brute_force_redis.get_failed"),
			slog.String("email", maskEmail(email)),
			slog.Any("error", err),
		)
		return 0, nil
	}
	return val, nil
}

// SetLockout sets a temporary lockout for an email for the specified duration.
// If an error occurs, it logs the failure and degrades gracefully returning nil error.
func (r *RedisBruteForceCache) SetLockout(ctx context.Context, email string, duration time.Duration) error {
	key := lockoutKeyPrefix + email
	expiryTime := time.Now().Add(duration)
	err := r.client.Set(ctx, key, expiryTime.Format(time.RFC3339), duration).Err()
	if err != nil {
		r.logger.ErrorContext(
			ctx, "failed to set lockout in Redis, degrading gracefully",
			slog.String("event", "brute_force_redis.set_lockout_failed"),
			slog.String("email", maskEmail(email)),
			slog.Any("error", err),
		)
		return nil
	}

	r.logger.InfoContext(
		ctx, "successfully locked account",
		slog.String("event", "brute_force_redis.set_lockout_success"),
		slog.String("email", maskEmail(email)),
		slog.Time("expiry", expiryTime),
	)
	return nil
}

// IsLocked checks if the email is currently locked out.
// It returns true and the lockout expiration time if locked, or false and a zero time if not locked.
// If an error occurs, it logs the failure and degrades gracefully (fail-open) returning false and a zero time.
func (r *RedisBruteForceCache) IsLocked(ctx context.Context, email string) (bool, time.Time, error) {
	key := lockoutKeyPrefix + email
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, time.Time{}, nil
		}
		r.logger.ErrorContext(
			ctx, "failed to query lockout status in Redis, degrading gracefully (fail-open)",
			slog.String("event", "brute_force_redis.is_locked_failed"),
			slog.String("email", maskEmail(email)),
			slog.Any("error", err),
		)
		return false, time.Time{}, nil
	}

	parsedTime, err := time.Parse(time.RFC3339, val)
	if err != nil {
		r.logger.WarnContext(
			ctx, "failed to parse lockout expiration time from Redis, calculating from TTL",
			slog.String("event", "brute_force_redis.parse_expiry_failed"),
			slog.String("email", maskEmail(email)),
			slog.Any("error", err),
		)
		ttl, ttlErr := r.client.TTL(ctx, key).Result()
		if ttlErr == nil && ttl > 0 {
			return true, time.Now().Add(ttl), nil
		}
		// Fallback default to fail-open (locked=false) if parsing and TTL both fail
		return false, time.Time{}, nil
	}

	return true, parsedTime, nil
}

// Reset clears both the failed attempts counter and the lockout state for an email.
// If an error occurs, it logs the failure and degrades gracefully returning nil error.
func (r *RedisBruteForceCache) Reset(ctx context.Context, email string) error {
	keyAttempts := failedAttemptsKeyPrefix + email
	keyLockout := lockoutKeyPrefix + email

	err := r.client.Del(ctx, keyAttempts, keyLockout).Err()
	if err != nil {
		r.logger.ErrorContext(
			ctx, "failed to reset brute force state in Redis, degrading gracefully",
			slog.String("event", "brute_force_redis.reset_failed"),
			slog.String("email", maskEmail(email)),
			slog.Any("error", err),
		)
		return nil
	}

	r.logger.InfoContext(
		ctx, "successfully reset brute force attempts and lockout state",
		slog.String("event", "brute_force_redis.reset_success"),
		slog.String("email", maskEmail(email)),
	)
	return nil
}
