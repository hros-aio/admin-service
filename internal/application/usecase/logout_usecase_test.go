package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSessionTokenRepo struct {
	mock.Mock
}

func (m *mockSessionTokenRepo) Save(ctx context.Context, token *domain.SessionToken) error {
	return m.Called(ctx, token).Error(0)
}

func (m *mockSessionTokenRepo) FindByToken(ctx context.Context, token string) (*domain.SessionToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SessionToken), args.Error(1)
}

func (m *mockSessionTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

func (m *mockSessionTokenRepo) DeleteByAdminID(ctx context.Context, adminID string) error {
	return m.Called(ctx, adminID).Error(0)
}

func (m *mockSessionTokenRepo) Revoke(ctx context.Context, token string, reason string) error {
	return m.Called(ctx, token, reason).Error(0)
}

type mockLogoutAuditLogger struct {
	mock.Mock
}

func (m *mockLogoutAuditLogger) LogLoginSuccess(ctx context.Context, userID string, email string) {
	m.Called(ctx, userID, email)
}

func (m *mockLogoutAuditLogger) LogLoginFailed(ctx context.Context, email string, reason string) {
	m.Called(ctx, email, reason)
}

func (m *mockLogoutAuditLogger) LogLogoutSuccess(ctx context.Context, userID string, token string) {
	m.Called(ctx, userID, token)
}

func TestLogoutUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sessionRepo := new(mockSessionTokenRepo)
		audit := new(mockLogoutAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "valid-session-token"
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "admin-id",
			RefreshToken: token,
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}

		sessionRepo.On("FindByToken", ctx, token).Return(session, nil).Once()
		sessionRepo.On("DeleteByToken", ctx, token).Return(nil).Once()
		audit.On("LogLogoutSuccess", ctx, "admin-id", token).Return().Once()

		err := uc.Execute(ctx, token)

		assert.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		sessionRepo := new(mockSessionTokenRepo)
		audit := new(mockLogoutAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		err := uc.Execute(ctx, "")

		assert.ErrorIs(t, err, domainErrors.ErrTokenNotFound)
	})

	t.Run("FindByTokenError", func(t *testing.T) {
		sessionRepo := new(mockSessionTokenRepo)
		audit := new(mockLogoutAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "valid-session-token"
		dbErr := errors.New("database connection failed")

		sessionRepo.On("FindByToken", ctx, token).Return(nil, dbErr).Once()

		err := uc.Execute(ctx, token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "find session token")
		assert.ErrorIs(t, err, dbErr)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("TokenNotFound", func(t *testing.T) {
		sessionRepo := new(mockSessionTokenRepo)
		audit := new(mockLogoutAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "unknown-session-token"

		sessionRepo.On("FindByToken", ctx, token).Return(nil, nil).Once()

		err := uc.Execute(ctx, token)

		assert.ErrorIs(t, err, domainErrors.ErrTokenNotFound)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("DeleteByTokenError", func(t *testing.T) {
		sessionRepo := new(mockSessionTokenRepo)
		audit := new(mockLogoutAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "valid-session-token"
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "admin-id",
			RefreshToken: token,
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}
		dbErr := errors.New("delete error")

		sessionRepo.On("FindByToken", ctx, token).Return(session, nil).Once()
		sessionRepo.On("DeleteByToken", ctx, token).Return(dbErr).Once()

		err := uc.Execute(ctx, token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete session token")
		assert.ErrorIs(t, err, dbErr)
		sessionRepo.AssertExpectations(t)
	})
}
