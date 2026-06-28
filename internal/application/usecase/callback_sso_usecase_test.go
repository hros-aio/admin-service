package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSSOClient struct {
	mock.Mock
}

func (m *mockSSOClient) ExchangeCode(ctx context.Context, provider string, code string) (*interfaces.SSOUserProfile, error) {
	args := m.Called(ctx, provider, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.SSOUserProfile), args.Error(1)
}

func TestCallbackSSOUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	input := CallbackSSOInput{
		Provider:  "google",
		Code:      "code-123",
		State:     "state-abc",
		IPAddress: "192.168.1.10",
		UserAgent: "Mozilla/5.0",
	}

	t.Run("success - callback completes login", func(t *testing.T) {
		stateCache := new(mockSSOStateCache)
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		uc := NewCallbackSSOUseCase(stateCache, userRepo, sessionRepo, tokens, ssoClient, audit)

		stateCache.On("VerifyAndConsumeState", ctx, "state-abc").Return("nonce-123", nil).Once()

		profile := &interfaces.SSOUserProfile{
			Email:      "sso-user@example.com",
			IdentityID: "sub-12345",
			Provider:   "google",
		}
		ssoClient.On("ExchangeCode", ctx, "google", "code-123").Return(profile, nil).Once()

		user := &domain.AdminUser{
			ID:    "user-uuid",
			Email: "sso-user@example.com",
			Name:  "SSO User",
		}
		userRepo.On("FindByEmailOrSSO", ctx, "sso-user@example.com", "google", "sub-12345").Return(user, nil).Once()

		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token-jwt", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("refresh-token-session", nil).Once()

		sessionRepo.On("Save", ctx, mock.MatchedBy(func(s *domain.SessionToken) bool {
			return s.AdminID == "user-uuid" && s.RefreshToken == "refresh-token-session"
		})).Return(nil).Once()

		audit.On("LogSSOSuccess", ctx, mock.MatchedBy(func(e events.SSOSuccessEvent) bool {
			return e.AdminID == "user-uuid" && e.Email == "sso-user@example.com" && e.Provider == "google"
		})).Return().Once()

		output, err := uc.Execute(ctx, input)

		require.NoError(t, err)
		assert.Equal(t, "access-token-jwt", output.AccessToken)
		assert.Equal(t, "refresh-token-session", output.RefreshToken)
		assert.Equal(t, "user-uuid", output.User.ID)
		assert.Equal(t, "sso-user@example.com", output.User.Email)

		stateCache.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		tokens.AssertExpectations(t)
		ssoClient.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("failure - empty provider", func(t *testing.T) {
		uc := NewCallbackSSOUseCase(nil, nil, nil, nil, nil, nil)
		_, err := uc.Execute(ctx, CallbackSSOInput{Provider: "", Code: "c", State: "s"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider cannot be empty")
	})

	t.Run("failure - empty code", func(t *testing.T) {
		uc := NewCallbackSSOUseCase(nil, nil, nil, nil, nil, nil)
		_, err := uc.Execute(ctx, CallbackSSOInput{Provider: "google", Code: "", State: "s"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization code cannot be empty")
	})

	t.Run("failure - empty state", func(t *testing.T) {
		uc := NewCallbackSSOUseCase(nil, nil, nil, nil, nil, nil)
		_, err := uc.Execute(ctx, CallbackSSOInput{Provider: "google", Code: "c", State: ""})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state parameter cannot be empty")
	})

	t.Run("failure - invalid state", func(t *testing.T) {
		stateCache := new(mockSSOStateCache)
		uc := NewCallbackSSOUseCase(stateCache, nil, nil, nil, nil, nil)

		stateCache.On("VerifyAndConsumeState", ctx, "state-abc").Return("", errors.New("state mismatch")).Once()

		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domainErrors.ErrInvalidSSOState))
		stateCache.AssertExpectations(t)
	})

	t.Run("failure - exchange code fails", func(t *testing.T) {
		stateCache := new(mockSSOStateCache)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		uc := NewCallbackSSOUseCase(stateCache, nil, nil, nil, ssoClient, audit)

		stateCache.On("VerifyAndConsumeState", ctx, "state-abc").Return("nonce-123", nil).Once()
		ssoClient.On("ExchangeCode", ctx, "google", "code-123").Return((*interfaces.SSOUserProfile)(nil), errors.New("idp error")).Once()

		audit.On("LogSSOFailed", ctx, mock.MatchedBy(func(e events.SSOFailedEvent) bool {
			return e.Provider == "google" && e.Reason == "code exchange failed"
		})).Return().Once()

		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exchange code")
		stateCache.AssertExpectations(t)
		ssoClient.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("failure - exchange code returns nil profile", func(t *testing.T) {
		stateCache := new(mockSSOStateCache)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		uc := NewCallbackSSOUseCase(stateCache, nil, nil, nil, ssoClient, audit)

		stateCache.On("VerifyAndConsumeState", ctx, "state-abc").Return("nonce-123", nil).Once()
		ssoClient.On("ExchangeCode", ctx, "google", "code-123").Return((*interfaces.SSOUserProfile)(nil), nil).Once()

		audit.On("LogSSOFailed", ctx, mock.MatchedBy(func(e events.SSOFailedEvent) bool {
			return e.Provider == "google" && e.Reason == "code exchange returned nil profile"
		})).Return().Once()

		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sso profile is nil")
		stateCache.AssertExpectations(t)
		ssoClient.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("failure - lookup user not found", func(t *testing.T) {
		stateCache := new(mockSSOStateCache)
		userRepo := new(mockUserRepo)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		uc := NewCallbackSSOUseCase(stateCache, userRepo, nil, nil, ssoClient, audit)

		stateCache.On("VerifyAndConsumeState", ctx, "state-abc").Return("nonce-123", nil).Once()

		profile := &interfaces.SSOUserProfile{
			Email:      "sso-user@example.com",
			IdentityID: "sub-123",
			Provider:   "google",
		}
		ssoClient.On("ExchangeCode", ctx, "google", "code-123").Return(profile, nil).Once()

		userRepo.On("FindByEmailOrSSO", ctx, "sso-user@example.com", "google", "sub-123").Return((*domain.AdminUser)(nil), domainErrors.ErrUserNotFound).Once()

		audit.On("LogSSOFailed", ctx, mock.MatchedBy(func(e events.SSOFailedEvent) bool {
			return e.Email == "sso-user@example.com" && e.Provider == "google" && e.Reason == "no account linked to this identity"
		})).Return().Once()

		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domainErrors.ErrNoAccountLinked))

		stateCache.AssertExpectations(t)
		ssoClient.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("failure - lookup user conflict", func(t *testing.T) {
		stateCache := new(mockSSOStateCache)
		userRepo := new(mockUserRepo)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		uc := NewCallbackSSOUseCase(stateCache, userRepo, nil, nil, ssoClient, audit)

		stateCache.On("VerifyAndConsumeState", ctx, "state-abc").Return("nonce-123", nil).Once()

		profile := &interfaces.SSOUserProfile{
			Email:      "sso-user@example.com",
			IdentityID: "sub-123",
			Provider:   "google",
		}
		ssoClient.On("ExchangeCode", ctx, "google", "code-123").Return(profile, nil).Once()

		userRepo.On("FindByEmailOrSSO", ctx, "sso-user@example.com", "google", "sub-123").Return((*domain.AdminUser)(nil), domainErrors.ErrIdentityConflict).Once()

		audit.On("LogSSOFailed", ctx, mock.MatchedBy(func(e events.SSOFailedEvent) bool {
			return e.Email == "sso-user@example.com" && e.Provider == "google" && e.Reason == "identity conflict: email and sso identity map to different users"
		})).Return().Once()

		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domainErrors.ErrIdentityConflict))

		stateCache.AssertExpectations(t)
		ssoClient.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("failure - save session token fails", func(t *testing.T) {
		stateCache := new(mockSSOStateCache)
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		uc := NewCallbackSSOUseCase(stateCache, userRepo, sessionRepo, tokens, ssoClient, audit)

		stateCache.On("VerifyAndConsumeState", ctx, "state-abc").Return("nonce-123", nil).Once()

		profile := &interfaces.SSOUserProfile{
			Email:      "sso-user@example.com",
			IdentityID: "sub-12345",
			Provider:   "google",
		}
		ssoClient.On("ExchangeCode", ctx, "google", "code-123").Return(profile, nil).Once()

		user := &domain.AdminUser{
			ID:    "user-uuid",
			Email: "sso-user@example.com",
			Name:  "SSO User",
		}
		userRepo.On("FindByEmailOrSSO", ctx, "sso-user@example.com", "google", "sub-12345").Return(user, nil).Once()

		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token-jwt", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("refresh-token-session", nil).Once()

		sessionRepo.On("Save", ctx, mock.Anything).Return(errors.New("db write failed")).Once()

		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "save session token")

		stateCache.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		tokens.AssertExpectations(t)
		ssoClient.AssertExpectations(t)
	})
}
