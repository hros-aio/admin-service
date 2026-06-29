package application

import (
	"log/slog"
	"testing"

	"github.com/hros/admin-service/internal/application/auth"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

type (
	dummyUserRepo      struct{ domain.AdminUserRepository }
	dummySessionRepo   struct{ domain.SessionTokenRepository }
	dummyTokenProvider struct{ auth.TokenProvider }
	dummyAuditLogger   struct{ authDomain.AuditLogger }
	dummyMFACache      struct{ interfaces.MFACache }
)

func TestModule_VerifyMFAUseCaseWiring(t *testing.T) {
	var verifiedUsecase *usecase.VerifyMFAUseCase

	app := fx.New(
		Module,
		fx.Provide(
			func() domain.AdminUserRepository { return &dummyUserRepo{} },
			func() domain.SessionTokenRepository { return &dummySessionRepo{} },
			func() auth.TokenProvider { return &dummyTokenProvider{} },
			func() authDomain.AuditLogger { return &dummyAuditLogger{} },
			func() interfaces.MFACache { return &dummyMFACache{} },
			slog.Default,
		),
		fx.Populate(&verifiedUsecase),
	)

	// We only need to check that fx was able to construct the app graph and resolve the usecase successfully.
	err := app.Err()
	assert.NoError(t, err)
	assert.NotNil(t, verifiedUsecase)
}
