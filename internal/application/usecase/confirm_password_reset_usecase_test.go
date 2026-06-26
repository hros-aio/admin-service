package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestConfirmPasswordResetUseCase_Execute(t *testing.T) {
	t.Parallel()

	// Sample data
	token := "valid-reset-token-12345"
	adminID := "user-uuid-999"
	email := "user@hros.io"
	validPassword := "SecureP@ss123"

	mockUser := &domain.AdminUser{
		ID:    adminID,
		Email: email,
	}

	tests := []struct {
		name          string
		input         ConfirmPasswordResetInput
		setupMocks    func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger)
		expectedError string
	}{
		{
			name: "Happy Path - Atomically consumes token, resets password, and wipes sessions",
			input: ConfirmPasswordResetInput{
				Token:     token,
				Password:  validPassword,
				IPAddress: "127.0.0.1",
				UserAgent: "Mozilla/5.0",
			},
			setupMocks: func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {
				ctx := mock.Anything
				// Atomically consume token
				cache.On("ConsumeToken", ctx, token).Return(adminID, nil).Once()
				ur.On("FindByID", ctx, adminID).Return(mockUser, nil).Once()

				// Expect password update with bcrypt cost 12
				ur.On("UpdatePassword", ctx, adminID, mock.MatchedBy(func(hash string) bool {
					cost, err := bcrypt.Cost([]byte(hash))
					return err == nil && cost == 12 && strings.HasPrefix(hash, "$2a$12$")
				})).Return(nil).Once()

				// Expect all sessions deleted
				sr.On("DeleteAllByAdminID", ctx, adminID).Return(nil).Once()

				// Expect audit log
				audit.On("LogPasswordResetCompleted", ctx, mock.MatchedBy(func(e events.PasswordResetCompletedEvent) bool {
					return e.Email == email &&
						e.IPAddress == "127.0.0.1" &&
						e.UserAgent == "Mozilla/5.0"
				})).Once()
			},
			expectedError: "",
		},
		{
			name: "Empty Token Input - Returns error",
			input: ConfirmPasswordResetInput{
				Token:    "",
				Password: validPassword,
			},
			setupMocks:    func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {},
			expectedError: "reset token cannot be empty",
		},
		{
			name: "Empty Password Input - Returns error",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: "",
			},
			setupMocks:    func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {},
			expectedError: "password cannot be empty",
		},
		{
			name: "Weak Password - Length too short",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: "Sh@1",
			},
			setupMocks:    func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {},
			expectedError: domainErrors.ErrPasswordWeak.Error(),
		},
		{
			name: "Weak Password - Multibyte Unicode runes (8 runes but 23 bytes) is rejected as weak",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: "🔑🔑🔑🔑🔑A1!", // 5 emojis + 3 characters = 8 runes (23 bytes)
			},
			setupMocks:    func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {},
			expectedError: domainErrors.ErrPasswordWeak.Error(),
		},
		{
			name: "Strong Password - Multibyte Unicode runes (10 runes, 28 bytes) is accepted",
			input: ConfirmPasswordResetInput{
				Token:     token,
				Password:  "🔑🔑🔑🔑🔑🔑🔑A1!", // 7 emojis + 3 characters = 10 runes (31 bytes)
				IPAddress: "127.0.0.1",
				UserAgent: "Mozilla/5.0",
			},
			setupMocks: func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {
				ctx := mock.Anything
				cache.On("ConsumeToken", ctx, token).Return(adminID, nil).Once()
				ur.On("FindByID", ctx, adminID).Return(mockUser, nil).Once()
				ur.On("UpdatePassword", ctx, adminID, mock.Anything).Return(nil).Once()
				sr.On("DeleteAllByAdminID", ctx, adminID).Return(nil).Once()
				audit.On("LogPasswordResetCompleted", ctx, mock.Anything).Once()
			},
			expectedError: "",
		},
		{
			name: "Weak Password - No Uppercase",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: "weakpassword@1",
			},
			setupMocks:    func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {},
			expectedError: domainErrors.ErrPasswordWeak.Error(),
		},
		{
			name: "Weak Password - No Digit",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: "WeakPassword@",
			},
			setupMocks:    func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {},
			expectedError: domainErrors.ErrPasswordWeak.Error(),
		},
		{
			name: "Weak Password - No Special Character",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: "WeakPassword123",
			},
			setupMocks:    func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {},
			expectedError: domainErrors.ErrPasswordWeak.Error(),
		},
		{
			name: "Expired/Missing Token in Cache - Returns ErrTokenExpired",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: validPassword,
			},
			setupMocks: func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {
				ctx := mock.Anything
				cache.On("ConsumeToken", ctx, token).Return("", domainErrors.ErrTokenExpired).Once()
			},
			expectedError: domainErrors.ErrTokenExpired.Error(),
		},
		{
			name: "Already Used Token - Returns ErrTokenUsed",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: validPassword,
			},
			setupMocks: func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {
				ctx := mock.Anything
				cache.On("ConsumeToken", ctx, token).Return("", domainErrors.ErrTokenUsed).Once()
			},
			expectedError: domainErrors.ErrTokenUsed.Error(),
		},
		{
			name: "Unexpected Cache Failure - Returns wrapped error and original err",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: validPassword,
			},
			setupMocks: func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {
				ctx := mock.Anything
				cache.On("ConsumeToken", ctx, token).Return("", errors.New("redis down connection timeout")).Once()
			},
			expectedError: "consume reset token: redis down connection timeout",
		},
		{
			name: "User Repo FindByID Error - Propagates failure",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: validPassword,
			},
			setupMocks: func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {
				ctx := mock.Anything
				cache.On("ConsumeToken", ctx, token).Return(adminID, nil).Once()
				ur.On("FindByID", ctx, adminID).Return(nil, errors.New("database failure")).Once()
			},
			expectedError: "find user: database failure",
		},
		{
			name: "User Repo UpdatePassword Error - Propagates failure",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: validPassword,
			},
			setupMocks: func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {
				ctx := mock.Anything
				cache.On("ConsumeToken", ctx, token).Return(adminID, nil).Once()
				ur.On("FindByID", ctx, adminID).Return(mockUser, nil).Once()
				ur.On("UpdatePassword", ctx, adminID, mock.Anything).Return(errors.New("write failure")).Once()
			},
			expectedError: "update password: write failure",
		},
		{
			name: "Session Repo DeleteAllByAdminID Error - Propagates failure",
			input: ConfirmPasswordResetInput{
				Token:    token,
				Password: validPassword,
			},
			setupMocks: func(ur *mockUserRepo, sr *mockSessionRepo, cache *mockPasswordResetCache, audit *mockAuditLogger) {
				ctx := mock.Anything
				cache.On("ConsumeToken", ctx, token).Return(adminID, nil).Once()
				ur.On("FindByID", ctx, adminID).Return(mockUser, nil).Once()
				ur.On("UpdatePassword", ctx, adminID, mock.Anything).Return(nil).Once()
				sr.On("DeleteAllByAdminID", ctx, adminID).Return(errors.New("db session wipe failure")).Once()
			},
			expectedError: "delete session tokens: db session wipe failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ur := new(mockUserRepo)
			sr := new(mockSessionRepo)
			cache := new(mockPasswordResetCache)
			audit := new(mockAuditLogger)

			tt.setupMocks(ur, sr, cache, audit)

			uc := NewConfirmPasswordResetUseCase(ur, sr, cache, audit)
			err := uc.Execute(context.Background(), tt.input)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			ur.AssertExpectations(t)
			sr.AssertExpectations(t)
			cache.AssertExpectations(t)
			audit.AssertExpectations(t)
		})
	}
}
