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

func TestRefreshSessionUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)

		uc := NewRefreshSessionUseCase(userRepo, sessionRepo, tokens, audit)

		tokenVal := "valid-refresh-token"
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: tokenVal,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		user := &domain.AdminUser{
			ID:     "user-123",
			Status: domain.AdminUserStatusActive,
		}

		newAccessToken := "new-access-token"
		newRefreshToken := "new-refresh-token"

		sessionRepo.On("FindByToken", ctx, tokenVal).Return(session, nil).Once()
		userRepo.On("FindByID", ctx, "user-123").Return(user, nil).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return(newAccessToken, nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return(newRefreshToken, nil).Once()
		sessionRepo.On("UpdateToken", ctx, mock.Anything).Return(nil).Once()
		audit.On("LogSessionRefreshed", ctx, "user-123").Return().Once()

		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: tokenVal})
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, newAccessToken, output.AccessToken)
		assert.Equal(t, newRefreshToken, output.RefreshToken)

		// Assert mocks met
		sessionRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		tokens.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("Empty Token Input", func(t *testing.T) {
		uc := NewRefreshSessionUseCase(nil, nil, nil, nil)
		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: ""})
		assert.ErrorIs(t, err, domainErrors.ErrInvalidRefreshToken)
		assert.Nil(t, output)
	})

	t.Run("Session Not Found", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		uc := NewRefreshSessionUseCase(nil, sessionRepo, nil, nil)

		tokenVal := "missing-token"
		sessionRepo.On("FindByToken", ctx, tokenVal).Return((*domain.SessionToken)(nil), errors.New("not found")).Once()

		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: tokenVal})
		assert.ErrorIs(t, err, domainErrors.ErrInvalidRefreshToken)
		assert.Nil(t, output)
	})

	t.Run("Session Revoked", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		uc := NewRefreshSessionUseCase(nil, sessionRepo, nil, nil)

		tokenVal := "revoked-token"
		revokedTime := time.Now()
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: tokenVal,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			RevokedAt:    &revokedTime,
		}

		sessionRepo.On("FindByToken", ctx, tokenVal).Return(session, nil).Once()

		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: tokenVal})
		assert.ErrorIs(t, err, domainErrors.ErrInvalidRefreshToken)
		assert.Nil(t, output)
	})

	t.Run("Session Expired", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		uc := NewRefreshSessionUseCase(nil, sessionRepo, nil, nil)

		tokenVal := "expired-token"
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: tokenVal,
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		}

		sessionRepo.On("FindByToken", ctx, tokenVal).Return(session, nil).Once()

		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: tokenVal})
		assert.ErrorIs(t, err, domainErrors.ErrTokenExpired)
		assert.Nil(t, output)
	})

	t.Run("User Not Found", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		uc := NewRefreshSessionUseCase(userRepo, sessionRepo, nil, nil)

		tokenVal := "valid-token-missing-user"
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "missing-user",
			RefreshToken: tokenVal,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		sessionRepo.On("FindByToken", ctx, tokenVal).Return(session, nil).Once()
		userRepo.On("FindByID", ctx, "missing-user").Return((*domain.AdminUser)(nil), domainErrors.ErrUserNotFound).Once()

		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: tokenVal})
		assert.ErrorIs(t, err, domainErrors.ErrInvalidRefreshToken)
		assert.Nil(t, output)
	})

	t.Run("User Inactive", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		uc := NewRefreshSessionUseCase(userRepo, sessionRepo, nil, nil)

		tokenVal := "valid-token-inactive-user"
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: tokenVal,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		user := &domain.AdminUser{
			ID:     "user-123",
			Status: "inactive",
		}

		sessionRepo.On("FindByToken", ctx, tokenVal).Return(session, nil).Once()
		userRepo.On("FindByID", ctx, "user-123").Return(user, nil).Once()

		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: tokenVal})
		assert.ErrorIs(t, err, domainErrors.ErrUserInactive)
		assert.Nil(t, output)
	})

	t.Run("User Locked", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		uc := NewRefreshSessionUseCase(userRepo, sessionRepo, nil, nil)

		tokenVal := "valid-token-locked-user"
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: tokenVal,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		lockedTime := time.Now().Add(1 * time.Hour)
		user := &domain.AdminUser{
			ID:          "user-123",
			Status:      domain.AdminUserStatusActive,
			LockedUntil: &lockedTime,
		}

		sessionRepo.On("FindByToken", ctx, tokenVal).Return(session, nil).Once()
		userRepo.On("FindByID", ctx, "user-123").Return(user, nil).Once()

		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: tokenVal})
		assert.ErrorIs(t, err, domainErrors.ErrUserLocked)
		assert.Nil(t, output)
	})

	t.Run("Token Provider Generation Error", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		uc := NewRefreshSessionUseCase(userRepo, sessionRepo, tokens, nil)

		tokenVal := "valid-token-provider-err"
		session := &domain.SessionToken{
			ID:           "session-id",
			AdminID:      "user-123",
			RefreshToken: tokenVal,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		user := &domain.AdminUser{
			ID:     "user-123",
			Status: domain.AdminUserStatusActive,
		}

		sessionRepo.On("FindByToken", ctx, tokenVal).Return(session, nil).Once()
		userRepo.On("FindByID", ctx, "user-123").Return(user, nil).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("", errors.New("jwt provider error")).Once()

		output, err := uc.Execute(ctx, RefreshInput{RefreshToken: tokenVal})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "jwt provider error")
		assert.Nil(t, output)
	})
}
