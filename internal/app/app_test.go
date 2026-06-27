package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestAppValidate(t *testing.T) {
	// Generate a real RSA private key for testing
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	pemString := string(pem.EncodeToMemory(pemBlock))

	// Set mandatory env vars for config.Load
	_ = os.Setenv("APP_NAME", "test-app")
	_ = os.Setenv("ENV", "test")
	_ = os.Setenv("PORT", "8080")
	_ = os.Setenv("DB_URL", "postgres://localhost")
	_ = os.Setenv("REDIS_URL", "redis://localhost")
	_ = os.Setenv("KAFKA_BROKERS", "localhost")
	_ = os.Setenv("JWT_PRIVATE_KEY", pemString)
	defer func() {
		_ = os.Unsetenv("APP_NAME")
		_ = os.Unsetenv("ENV")
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DB_URL")
		_ = os.Unsetenv("REDIS_URL")
		_ = os.Unsetenv("KAFKA_BROKERS")
		_ = os.Unsetenv("JWT_PRIVATE_KEY")
	}()

	err = fx.ValidateApp(Module)
	require.NoError(t, err)
}

func TestLockoutNotifierBinding(t *testing.T) {
	// Generate a real RSA private key for testing
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	pemString := string(pem.EncodeToMemory(pemBlock))

	// Set mandatory env vars for config.Load
	_ = os.Setenv("APP_NAME", "test-app")
	_ = os.Setenv("ENV", "test")
	_ = os.Setenv("PORT", "8080")
	_ = os.Setenv("DB_URL", "postgres://localhost")
	_ = os.Setenv("REDIS_URL", "redis://localhost")
	_ = os.Setenv("KAFKA_BROKERS", "localhost")
	_ = os.Setenv("JWT_PRIVATE_KEY", pemString)
	defer func() {
		_ = os.Unsetenv("APP_NAME")
		_ = os.Unsetenv("ENV")
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DB_URL")
		_ = os.Unsetenv("REDIS_URL")
		_ = os.Unsetenv("KAFKA_BROKERS")
		_ = os.Unsetenv("JWT_PRIVATE_KEY")
	}()

	// Use fx.ValidateApp with an Invoke that requests interfaces.LockoutNotifier.
	// This verifies the binding resolves correctly in the graph without starting the app (avoiding actual DB/Redis connections).
	err = fx.ValidateApp(
		Module,
		fx.Invoke(func(notifier interfaces.LockoutNotifier) {
			require.NotNil(t, notifier)
		}),
	)
	require.NoError(t, err)
}

func TestInitiateSSOUseCaseWiring(t *testing.T) {
	// Generate a real RSA private key for testing
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	pemString := string(pem.EncodeToMemory(pemBlock))

	// Set mandatory env vars for config.Load
	_ = os.Setenv("APP_NAME", "test-app")
	_ = os.Setenv("ENV", "test")
	_ = os.Setenv("PORT", "8080")
	_ = os.Setenv("DB_URL", "postgres://localhost")
	_ = os.Setenv("REDIS_URL", "redis://localhost")
	_ = os.Setenv("KAFKA_BROKERS", "localhost")
	_ = os.Setenv("JWT_PRIVATE_KEY", pemString)
	defer func() {
		_ = os.Unsetenv("APP_NAME")
		_ = os.Unsetenv("ENV")
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DB_URL")
		_ = os.Unsetenv("REDIS_URL")
		_ = os.Unsetenv("KAFKA_BROKERS")
		_ = os.Unsetenv("JWT_PRIVATE_KEY")
	}()

	err = fx.ValidateApp(
		Module,
		fx.Invoke(func(uc *usecase.InitiateSSOUseCase) {
			require.NotNil(t, uc)
		}),
	)
	require.NoError(t, err)
}

