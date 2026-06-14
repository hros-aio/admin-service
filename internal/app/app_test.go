package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestAppValidate(t *testing.T) {
	// Set mandatory env vars for config.Load
	_ = os.Setenv("APP_NAME", "test-app")
	_ = os.Setenv("ENV", "test")
	_ = os.Setenv("PORT", "8080")
	_ = os.Setenv("DB_URL", "postgres://localhost")
	_ = os.Setenv("REDIS_URL", "redis://localhost")
	_ = os.Setenv("KAFKA_BROKERS", "localhost")
	defer func() {
		_ = os.Unsetenv("APP_NAME")
		_ = os.Unsetenv("ENV")
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DB_URL")
		_ = os.Unsetenv("REDIS_URL")
		_ = os.Unsetenv("KAFKA_BROKERS")
	}()

	err := fx.ValidateApp(Module)
	require.NoError(t, err)
}
