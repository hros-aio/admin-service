package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPasswordResetCache struct {
	mock.Mock
}

func (m *mockPasswordResetCache) StoreToken(ctx context.Context, token string, adminID string, ttl time.Duration) error {
	return m.Called(ctx, token, adminID, ttl).Error(0)
}

func (m *mockPasswordResetCache) ConsumeToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

type mockPasswordResetNotifier struct {
	mock.Mock
}

func (m *mockPasswordResetNotifier) PublishPasswordResetEmail(ctx context.Context, event events.EmailSendEvent) error {
	return m.Called(ctx, event).Error(0)
}

func TestRequestPasswordResetUseCase_Execute(t *testing.T) {
	t.Parallel()

	// Sample data
	registeredEmail := "registered@example.com"
	unregisteredEmail := "unregistered@example.com"
	mockAdminUser := &domain.AdminUser{
		ID:    "user-uuid-123",
		Email: registeredEmail,
	}

	tests := []struct {
		name          string
		input         RequestPasswordResetInput
		setupMocks    func(ur *mockUserRepo, cache *mockPasswordResetCache, audit *mockAuditLogger, notifier *mockPasswordResetNotifier)
		expectedError string
	}{
		{
			name: "Happy Path - Registered Email triggers reset flow",
			input: RequestPasswordResetInput{
				Email:     registeredEmail,
				IPAddress: "127.0.0.1",
				UserAgent: "Mozilla/5.0",
			},
			setupMocks: func(ur *mockUserRepo, cache *mockPasswordResetCache, audit *mockAuditLogger, notifier *mockPasswordResetNotifier) {
				ctx := mock.Anything
				ur.On("FindByEmail", ctx, registeredEmail).Return(mockAdminUser, nil).Once()

				var capturedToken string
				// Cache token expectation: length 64 hex, matching user ID, 60m TTL
				cache.On("StoreToken", ctx, mock.MatchedBy(func(token string) bool {
					capturedToken = token
					return len(token) == 64
				}), mockAdminUser.ID, 60*time.Minute).Return(nil).Once()

				// Audit log expectation
				audit.On("LogPasswordResetRequested", ctx, mock.MatchedBy(func(e events.PasswordResetRequestedEvent) bool {
					return e.Email == registeredEmail &&
						e.Token == capturedToken &&
						e.IPAddress == "127.0.0.1" &&
						e.UserAgent == "Mozilla/5.0"
				})).Once()

				// Notifier expectation
				notifier.On("PublishPasswordResetEmail", ctx, mock.MatchedBy(func(event events.EmailSendEvent) bool {
					return event.To == registeredEmail &&
						event.Subject == "Reset your password" &&
						event.Template == "password_reset_request" &&
						event.TemplateData["token"] == capturedToken &&
						event.TemplateData["email"] == registeredEmail
				})).Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name: "Unregistered Email - Prevents enumeration by returning success immediately",
			input: RequestPasswordResetInput{
				Email: unregisteredEmail,
			},
			setupMocks: func(ur *mockUserRepo, cache *mockPasswordResetCache, audit *mockAuditLogger, notifier *mockPasswordResetNotifier) {
				ctx := mock.Anything
				ur.On("FindByEmail", ctx, unregisteredEmail).Return(nil, domainErrors.ErrUserNotFound).Once()
				// No other mocks should be called
			},
			expectedError: "",
		},
		{
			name:  "Empty Email Input - Returns error",
			input: RequestPasswordResetInput{Email: ""},
			setupMocks: func(ur *mockUserRepo, cache *mockPasswordResetCache, audit *mockAuditLogger, notifier *mockPasswordResetNotifier) {
			},
			expectedError: "email cannot be empty",
		},
		{
			name: "User Repo DB Error - Propagates repository error",
			input: RequestPasswordResetInput{
				Email: registeredEmail,
			},
			setupMocks: func(ur *mockUserRepo, cache *mockPasswordResetCache, audit *mockAuditLogger, notifier *mockPasswordResetNotifier) {
				ctx := mock.Anything
				ur.On("FindByEmail", ctx, registeredEmail).Return(nil, errors.New("db query failed")).Once()
			},
			expectedError: "find user: db query failed",
		},
		{
			name: "Cache Store Error - Propagates cache store failure",
			input: RequestPasswordResetInput{
				Email: registeredEmail,
			},
			setupMocks: func(ur *mockUserRepo, cache *mockPasswordResetCache, audit *mockAuditLogger, notifier *mockPasswordResetNotifier) {
				ctx := mock.Anything
				ur.On("FindByEmail", ctx, registeredEmail).Return(mockAdminUser, nil).Once()
				cache.On("StoreToken", ctx, mock.Anything, mockAdminUser.ID, 60*time.Minute).
					Return(errors.New("redis error")).Once()
			},
			expectedError: "store reset token: redis error",
		},
		{
			name: "Notifier Publish Error - Propagates Kafka publisher failure and rolls back token in cache",
			input: RequestPasswordResetInput{
				Email:     registeredEmail,
				IPAddress: "192.168.1.1",
				UserAgent: "Safari",
			},
			setupMocks: func(ur *mockUserRepo, cache *mockPasswordResetCache, audit *mockAuditLogger, notifier *mockPasswordResetNotifier) {
				ctx := mock.Anything
				ur.On("FindByEmail", ctx, registeredEmail).Return(mockAdminUser, nil).Once()

				var capturedToken string
				cache.On("StoreToken", ctx, mock.MatchedBy(func(token string) bool {
					capturedToken = token
					return len(token) == 64
				}), mockAdminUser.ID, 60*time.Minute).Return(nil).Once()

				audit.On("LogPasswordResetRequested", ctx, mock.MatchedBy(func(e events.PasswordResetRequestedEvent) bool {
					return e.Email == registeredEmail && e.Token == capturedToken
				})).Once()

				notifier.On("PublishPasswordResetEmail", ctx, mock.Anything).
					Return(errors.New("kafka connection reset")).Once()

				// Verify compensation consume is called
				cache.On("ConsumeToken", ctx, mock.MatchedBy(func(token string) bool {
					return token == capturedToken
				})).Return("", nil).Once()
			},
			expectedError: "publish password reset email: kafka connection reset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ur := new(mockUserRepo)
			cache := new(mockPasswordResetCache)
			audit := new(mockAuditLogger)
			notifier := new(mockPasswordResetNotifier)

			tt.setupMocks(ur, cache, audit, notifier)

			uc := NewRequestPasswordResetUseCase(ur, cache, audit, notifier)
			err := uc.Execute(context.Background(), tt.input)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			ur.AssertExpectations(t)
			cache.AssertExpectations(t)
			audit.AssertExpectations(t)
			notifier.AssertExpectations(t)
		})
	}
}
