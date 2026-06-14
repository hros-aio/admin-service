package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Helper to set env vars
	setEnv := func() {
		_ = os.Setenv("APP_NAME", "hros-admin")
		_ = os.Setenv("ENV", "local")
		_ = os.Setenv("PORT", "8080")
		_ = os.Setenv("DB_URL", "postgres://localhost:5432")
		_ = os.Setenv("REDIS_URL", "redis://localhost:6379")
		_ = os.Setenv("KAFKA_BROKERS", "localhost:9092")
	}

	// Helper to clear env vars
	clearEnv := func() {
		_ = os.Unsetenv("APP_NAME")
		_ = os.Unsetenv("ENV")
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DB_URL")
		_ = os.Unsetenv("REDIS_URL")
		_ = os.Unsetenv("KAFKA_BROKERS")
	}

	t.Run("success", func(t *testing.T) {
		setEnv()
		defer clearEnv()

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "hros-admin", cfg.AppName)
		require.Equal(t, "local", cfg.Env)
		require.Equal(t, 8080, cfg.Port)
		require.Equal(t, "info", cfg.LogLevel) // Default value
	})

	t.Run("missing_required", func(t *testing.T) {
		clearEnv()
		_, err := Load()
		require.Error(t, err)
	})

	t.Run("invalid_port", func(t *testing.T) {
		setEnv()
		defer clearEnv()
		_ = os.Setenv("PORT", "0")

		_, err := Load()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid port")
	})
}
