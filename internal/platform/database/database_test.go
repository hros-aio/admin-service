package database

import (
	"log/slog"
	"os"
	"testing"

	"github.com/hros/admin-service/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestNewDatabase(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := &config.Config{
		DBURL: "postgres://localhost:5432/test?sslmode=disable",
	}

	t.Run("wiring", func(t *testing.T) {
		err := fx.ValidateApp(
			fx.Provide(func() *config.Config { return cfg }),
			fx.Provide(func() *slog.Logger { return log }),
			fx.Provide(NewDatabase),
		)
		assert.NoError(t, err)
	})
}
