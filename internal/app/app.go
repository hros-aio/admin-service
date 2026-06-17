// Package app provides the application composition and lifecycle management.
package app

import (
	"log/slog"

	"github.com/hros/admin-service/internal/config"
	authRepo "github.com/hros/admin-service/internal/infrastructure/repository/auth"
	"github.com/hros/admin-service/internal/platform/database"
	"github.com/hros/admin-service/internal/platform/http"
	"github.com/hros/admin-service/internal/platform/kafka"
	"github.com/hros/admin-service/internal/platform/logger"
	"github.com/hros/admin-service/internal/platform/redis"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Module is the root Fx module for the application.
var Module = fx.Options(
	// Provide configuration
	fx.Provide(config.Load),

	// Provide logger
	fx.Provide(func(cfg *config.Config) *slog.Logger {
		return logger.New(cfg.LogLevel)
	}),

	// Infrastructure
	fx.Provide(database.NewDatabase),
	fx.Provide(database.NewTxManager),
	fx.Provide(authRepo.NewGormAdminUserRepository),
	fx.Provide(redis.NewRedisClient),
	fx.Provide(kafka.NewKafkaProducer),
	fx.Provide(kafka.NewKafkaConsumerGroup),

	// Adapters/Handlers
	fx.Provide(http.NewHealthHandler),
	fx.Provide(http.NewServer),

	// Invokes
	fx.Invoke(func(_ *echo.Echo) {}),

	// Configure Fx logging to use our structured logger
	fx.WithLogger(func(log *slog.Logger) fxevent.Logger {
		return &fxevent.SlogLogger{Logger: log}
	}),
)

// New initializes the Fx application.
func New() *fx.App {
	return fx.New(Module)
}
