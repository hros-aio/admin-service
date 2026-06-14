// Package redis provides Redis client initialization and utilities.
package redis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hros/admin-service/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// NewRedisClient initializes the Redis client.
func NewRedisClient(cfg *config.Config, _ *slog.Logger, lc fx.Lifecycle) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return client.Close()
		},
	})

	return client, nil
}
