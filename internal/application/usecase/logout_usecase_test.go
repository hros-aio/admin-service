package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLogoutSessionRepo struct {
	mock.Mock
}

func (m *mockLogoutSessionRepo) Save(ctx context.Context, token *domain.SessionToken) error {
	return m.Called(ctx, token).Error(0)
}

func (m *mockLogoutSessionRepo) FindByToken(ctx context.Context, token string) (*domain.SessionToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SessionToken), args.Error(1)
}

func (m *mockLogoutSessionRepo) DeleteByToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

func (m *mockLogoutSessionRepo) DeleteByAdminID(ctx context.Context, adminID string) error {
	return m.Called(ctx, adminID).Error(0)
}

func (m *mockLogoutSessionRepo) Revoke(ctx context.Context, token string, reason string) error {
	return m.Called(ctx, token, reason).Error(0)
}

func TestLogoutUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sessionRepo := new(mockLogoutSessionRepo)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "active-session-token"
		input := LogoutInput{Token: token}

		sessionRepo.On("DeleteByToken", ctx, token).Return(nil).Once()
		audit.On("LogLogoutSuccess", ctx, token).Return().Once()

		err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("TokenEmpty", func(t *testing.T) {
		sessionRepo := new(mockLogoutSessionRepo)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		input := LogoutInput{Token: ""}

		err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, domainErrors.ErrTokenNotFound)
		assert.Contains(t, err.Error(), "session token is empty")
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("TokenNotFound", func(t *testing.T) {
		sessionRepo := new(mockLogoutSessionRepo)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "non-existent-token"
		input := LogoutInput{Token: token}

		sessionRepo.On("DeleteByToken", ctx, token).Return(domainErrors.ErrTokenNotFound).Once()

		err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, domainErrors.ErrTokenNotFound)
		assert.Contains(t, err.Error(), "delete session token")
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		sessionRepo := new(mockLogoutSessionRepo)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "some-token"
		input := LogoutInput{Token: token}
		dbErr := errors.New("db connection failure")

		sessionRepo.On("DeleteByToken", ctx, token).Return(dbErr).Once()

		err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, dbErr)
		assert.Contains(t, err.Error(), "delete session token")
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})
}
