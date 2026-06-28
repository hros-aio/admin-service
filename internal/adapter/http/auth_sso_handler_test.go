package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/adapter/http/auth/dto"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
	sharedErrors "github.com/hros/admin-service/internal/shared/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSSOStateCache struct {
	mock.Mock
}

func (m *mockSSOStateCache) StoreState(ctx context.Context, state string, nonce string, ttl time.Duration) error {
	return m.Called(ctx, state, nonce, ttl).Error(0)
}

func (m *mockSSOStateCache) VerifyAndConsumeState(ctx context.Context, state string) (string, error) {
	args := m.Called(ctx, state)
	return args.String(0), args.Error(1)
}

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
func TestAuthSSOHandler_Initiate(t *testing.T) {
	e := echo.New()

	providers := map[string]usecase.SSOProviderConfig{
		"google": {
			ClientID:    "google-client-id",
			RedirectURL: "https://hros.io/callback",
			AuthURL:     "https://accounts.google.com/o/oauth2/auth",
			Scopes:      []string{"openid", "email"},
		},
	}

	t.Run("success - initiate SSO redirect", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		cache.On("StoreState", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 10*time.Minute).Return(nil).Once()

		initUC := usecase.NewInitiateSSOUseCase(cache, providers)
		h := NewAuthSSOHandler(initUC, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/sso/initiate?provider=google", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Initiate(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, rec.Code)

		loc := rec.Header().Get("Location")
		assert.Contains(t, loc, "https://accounts.google.com/o/oauth2/auth")
		assert.Contains(t, loc, "client_id=google-client-id")
		assert.Contains(t, loc, "redirect_uri=https%3A%2F%2Fhros.io%2Fcallback")
		assert.Contains(t, loc, "state=")
		assert.Contains(t, loc, "nonce=")

		cache.AssertExpectations(t)
	})

	t.Run("failure - missing provider", func(t *testing.T) {
		h := NewAuthSSOHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/sso/initiate", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Initiate(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", resp.Code)
	})

	t.Run("failure - unsupported provider", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		initUC := usecase.NewInitiateSSOUseCase(cache, providers)
		h := NewAuthSSOHandler(initUC, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/sso/initiate?provider=github", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Initiate(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "bad_request", resp.Code)
	})
}

func TestAuthSSOHandler_Callback(t *testing.T) {
	e := echo.New()

	t.Run("success - callback returns JSON when Accept application/json", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		callbackUC := usecase.NewCallbackSSOUseCase(cache, userRepo, sessionRepo, tokens, ssoClient, audit)
		h := NewAuthSSOHandler(nil, callbackUC)

		cache.On("VerifyAndConsumeState", mock.Anything, "state-abc").Return("nonce-123", nil).Once()

		profile := &interfaces.SSOUserProfile{
			Email:      "sso-user@example.com",
			IdentityID: "sub-12345",
			Provider:   "google",
		}
		ssoClient.On("ExchangeCode", mock.Anything, "google", "code-123").Return(profile, nil).Once()

		user := &domain.AdminUser{
			ID:    "user-uuid",
			Email: "sso-user@example.com",
			Name:  "SSO User",
		}
		userRepo.On("FindByEmailOrSSO", mock.Anything, "sso-user@example.com", "google", "sub-12345").Return(user, nil).Once()

		tokens.On("GenerateAccessToken", mock.Anything, user, 15*time.Minute).Return("access-token-jwt", nil).Once()
		tokens.On("GenerateRefreshToken", mock.Anything).Return("refresh-token-session", nil).Once()

		sessionRepo.On("Save", mock.Anything, mock.MatchedBy(func(s *domain.SessionToken) bool {
			return s.AdminID == "user-uuid" && s.RefreshToken == "refresh-token-session"
		})).Return(nil).Once()

		audit.On("LogSSOSuccess", mock.Anything, mock.AnythingOfType("events.SSOSuccessEvent")).Return().Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/sso/callback?code=code-123&state=state-abc&provider=google", nil)
		req.Header.Set("Accept", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Callback(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.LoginResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "access-token-jwt", resp.AccessToken)
		assert.Equal(t, "refresh-token-session", resp.RefreshToken)

		// Assert cookie
		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "refresh_token", cookies[0].Name)
		assert.Equal(t, "refresh-token-session", cookies[0].Value)
		assert.True(t, cookies[0].HttpOnly)

		cache.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		tokens.AssertExpectations(t)
		ssoClient.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("success - callback redirects when Accept text/html", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		callbackUC := usecase.NewCallbackSSOUseCase(cache, userRepo, sessionRepo, tokens, ssoClient, audit)
		h := NewAuthSSOHandler(nil, callbackUC)

		cache.On("VerifyAndConsumeState", mock.Anything, "state-abc").Return("nonce-123", nil).Once()

		profile := &interfaces.SSOUserProfile{
			Email:      "sso-user@example.com",
			IdentityID: "sub-12345",
			Provider:   "google",
		}
		ssoClient.On("ExchangeCode", mock.Anything, "google", "code-123").Return(profile, nil).Once()

		user := &domain.AdminUser{
			ID:    "user-uuid",
			Email: "sso-user@example.com",
			Name:  "SSO User",
		}
		userRepo.On("FindByEmailOrSSO", mock.Anything, "sso-user@example.com", "google", "sub-12345").Return(user, nil).Once()

		tokens.On("GenerateAccessToken", mock.Anything, user, 15*time.Minute).Return("access-token-jwt", nil).Once()
		tokens.On("GenerateRefreshToken", mock.Anything).Return("refresh-token-session", nil).Once()

		sessionRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

		audit.On("LogSSOSuccess", mock.Anything, mock.Anything).Return().Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/sso/callback?code=code-123&state=state-abc", nil)
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Callback(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "/dashboard", rec.Header().Get("Location"))

		cache.AssertExpectations(t)
	})

	t.Run("failure - invalid state", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		callbackUC := usecase.NewCallbackSSOUseCase(cache, nil, nil, nil, nil, nil)
		h := NewAuthSSOHandler(nil, callbackUC)

		cache.On("VerifyAndConsumeState", mock.Anything, "state-abc").Return("", errors.New("state mismatch")).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/sso/callback?code=code-123&state=state-abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Callback(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_SSO_STATE", resp.Code)
	})

	t.Run("failure - unlinked account", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		userRepo := new(mockUserRepo)
		ssoClient := new(mockSSOClient)
		audit := new(mockAuditLogger)

		callbackUC := usecase.NewCallbackSSOUseCase(cache, userRepo, nil, nil, ssoClient, audit)
		h := NewAuthSSOHandler(nil, callbackUC)

		cache.On("VerifyAndConsumeState", mock.Anything, "state-abc").Return("nonce-123", nil).Once()

		profile := &interfaces.SSOUserProfile{
			Email:      "sso-user@example.com",
			IdentityID: "sub-12345",
			Provider:   "google",
		}
		ssoClient.On("ExchangeCode", mock.Anything, "google", "code-123").Return(profile, nil).Once()

		userRepo.On("FindByEmailOrSSO", mock.Anything, "sso-user@example.com", "google", "sub-12345").Return((*domain.AdminUser)(nil), domainErrors.ErrUserNotFound).Once()

		audit.On("LogSSOFailed", mock.Anything, mock.MatchedBy(func(e events.SSOFailedEvent) bool {
			return e.Email == "sso-user@example.com" && e.Reason == "no account linked to this identity"
		})).Return().Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/sso/callback?code=code-123&state=state-abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.Callback(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var resp sharedErrors.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "NO_ACCOUNT_LINKED", resp.Code)
		assert.Equal(t, "No admin account linked to this identity", resp.Message)
	})
}
