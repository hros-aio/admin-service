package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogoutUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "valid-refresh-token"
		input := LogoutInput{RefreshToken: token}

		sessionRepo.On("DeleteByToken", ctx, token).Return(nil).Once()
		audit.On("LogLogoutSuccess", ctx, token).Return().Once()

		err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		input := LogoutInput{RefreshToken: ""}

		err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token cannot be empty")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		sessionRepo := new(mockSessionRepo)
		audit := new(mockAuditLogger)
		uc := NewLogoutUseCase(sessionRepo, audit)

		token := "some-token"
		input := LogoutInput{RefreshToken: token}
		dbErr := errors.New("database connection lost")

		sessionRepo.On("DeleteByToken", ctx, token).Return(dbErr).Once()

		err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
		sessionRepo.AssertExpectations(t)
		audit.AssertNotCalled(t, "LogLogoutSuccess", mock.Anything, mock.Anything)
	})
}
