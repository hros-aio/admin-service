package http

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hros/admin-service/internal/config"
	"github.com/hros/admin-service/internal/shared/middleware"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
)

// NewServer initializes the Echo HTTP server.
func NewServer(cfg *config.Config, log *slog.Logger, lc fx.Lifecycle, health *HealthHandler) *echo.Echo {
	e := echo.New()

	// Standard middleware
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.RequestID())
	e.Use(middleware.Logger(log))

	// Routes
	e.GET("/health", health.Check)

	// Swagger UI
	e.Static("/openapi", "api")
	e.File("/docs", "docs/openapi/index.html")

	// Lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				addr := fmt.Sprintf(":%d", cfg.Port)
				if err := e.Start(addr); err != nil {
					log.Error("failed to start server", slog.Any("error", err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return e.Shutdown(ctx)
		},
	})

	return e
}
