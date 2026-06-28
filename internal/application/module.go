package application

import (
	"github.com/hros/admin-service/internal/application/usecase"
	"go.uber.org/fx"
)

// Module is the Fx module for application use cases.
var Module = fx.Module("application",
	fx.Provide(
		usecase.NewLoginUseCase,
		usecase.NewLogoutUseCase,
		usecase.NewRefreshSessionUseCase,
		usecase.NewVerifyMFAUseCase,
		usecase.NewRequestPasswordResetUseCase,
		usecase.NewConfirmPasswordResetUseCase,
		usecase.NewAcceptInviteUseCase,
		usecase.NewInitiateSSOUseCase,
		usecase.NewCallbackSSOUseCase,
		usecase.NewGenerateBiometricChallengeUseCase,
		usecase.NewVerifyBiometricUseCase,
	),
)
