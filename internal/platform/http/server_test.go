package http

import (
	"log/slog"
	"os"
	"testing"

	"github.com/hros/admin-service/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewServer(t *testing.T) {
	// Mock config
	cfg := &config.Config{
		Port: 8081,
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := fxtest.New(
		t,
		fx.Provide(func() *config.Config { return cfg }),
		fx.Provide(func() *slog.Logger { return log }),
		fx.Provide(func() *HealthHandler { return &HealthHandler{} }),
		fx.Provide(NewServer),
		fx.Invoke(func(e *echo.Echo) {
			assert.NotNil(t, e)
		}),
	)
	app.RequireStart()
	app.RequireStop()
}
