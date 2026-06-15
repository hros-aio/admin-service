package redis

import (
	"log/slog"
	"os"
	"testing"

	"github.com/hros/admin-service/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewRedisClient(t *testing.T) {
	cfg := &config.Config{
		RedisURL: "redis://localhost:6379/0",
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fxtest.New(t,
		fx.Provide(func() *config.Config { return cfg }),
		fx.Provide(func() *slog.Logger { return log }),
		fx.Provide(NewRedisClient),
		fx.Invoke(func(client *redis.Client) {
			assert.NotNil(t, client)
		}),
	)
}
