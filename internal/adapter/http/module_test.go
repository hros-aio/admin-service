package http

import (
	"context"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestHTTPModuleWiring(t *testing.T) {
	app := fx.New(
		Module,
		fx.Provide(
			func() *echo.Echo {
				return echo.New()
			},
			func() *usecase.LoginUseCase {
				return nil
			},
			func() *usecase.LogoutUseCase {
				return nil
			},
			func() *usecase.RefreshSessionUseCase {
				return nil
			},
			func() *usecase.VerifyMFAUseCase {
				return nil
			},
			func() *usecase.RequestPasswordResetUseCase {
				return nil
			},
			func() *usecase.ConfirmPasswordResetUseCase {
				return nil
			},
			func() *usecase.InitiateSSOUseCase {
				return nil
			},
			func() *usecase.CallbackSSOUseCase {
				return nil
			},
			func() *usecase.GenerateBiometricChallengeUseCase {
				return nil
			},
			func() *usecase.VerifyBiometricUseCase {
				return nil
			},
		),
	)

	// Validate constructor graph
	err := app.Err()
	assert.NoError(t, err)

	// Assert the app starts successfully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.Start(ctx)
	assert.NoError(t, err)

	err = app.Stop(ctx)
	assert.NoError(t, err)
}
