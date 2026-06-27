package usecase

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestInitiateSSOUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	providers := map[string]SSOProviderConfig{
		"google": {
			ClientID:    "google-client-id",
			RedirectURL: "https://example.com/v1/auth/sso/callback",
			AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
			Scopes:      []string{"openid", "email", "profile"},
		},
		"okta": {
			ClientID:    "okta-client-id",
			RedirectURL: "https://example.com/v1/auth/sso/callback",
			AuthURL:     "https://okta.com/oauth2/v1/authorize",
			Scopes:      nil, // triggers fallback scopes
		},
		"malformed": {
			ClientID:    "malformed-client-id",
			RedirectURL: "https://example.com/v1/auth/sso/callback",
			AuthURL:     "://invalid-url",
		},
		"incomplete": {
			ClientID: "",
		},
	}

	t.Run("success - google login initiation", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		uc := NewInitiateSSOUseCase(cache, providers)

		cache.On("StoreState", ctx, mock.Anything, mock.Anything, 10*time.Minute).Return(nil).Once()

		input := InitiateSSOInput{Provider: "Google"} // verify case insensitivity
		output, err := uc.Execute(ctx, input)

		require.NoError(t, err)
		assert.NotEmpty(t, output.RedirectURL)

		parsedURL, err := url.Parse(output.RedirectURL)
		require.NoError(t, err)
		assert.Equal(t, "accounts.google.com", parsedURL.Host)
		assert.Equal(t, "/o/oauth2/v2/auth", parsedURL.Path)

		q := parsedURL.Query()
		assert.Equal(t, "google-client-id", q.Get("client_id"))
		assert.Equal(t, "https://example.com/v1/auth/sso/callback", q.Get("redirect_uri"))
		assert.Equal(t, "code", q.Get("response_type"))
		assert.Equal(t, "openid email profile", q.Get("scope"))
		assert.NotEmpty(t, q.Get("state"))
		assert.NotEmpty(t, q.Get("nonce"))

		cache.AssertExpectations(t)
	})

	t.Run("success - okta login with default scope fallback", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		uc := NewInitiateSSOUseCase(cache, providers)

		cache.On("StoreState", ctx, mock.Anything, mock.Anything, 10*time.Minute).Return(nil).Once()

		input := InitiateSSOInput{Provider: "okta"}
		output, err := uc.Execute(ctx, input)

		require.NoError(t, err)
		assert.NotEmpty(t, output.RedirectURL)

		parsedURL, err := url.Parse(output.RedirectURL)
		require.NoError(t, err)
		assert.Equal(t, "openid email profile", parsedURL.Query().Get("scope"))

		cache.AssertExpectations(t)
	})

	t.Run("failure - empty provider name", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		uc := NewInitiateSSOUseCase(cache, providers)

		input := InitiateSSOInput{Provider: ""}
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider cannot be empty")
	})

	t.Run("failure - unsupported provider", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		uc := NewInitiateSSOUseCase(cache, providers)

		input := InitiateSSOInput{Provider: "github"}
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported or unconfigured provider")
	})

	t.Run("failure - incomplete provider configuration", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		uc := NewInitiateSSOUseCase(cache, providers)

		input := InitiateSSOInput{Provider: "incomplete"}
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration for provider")
	})

	t.Run("failure - malformed provider auth URL", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		uc := NewInitiateSSOUseCase(cache, providers)

		cache.On("StoreState", ctx, mock.Anything, mock.Anything, 10*time.Minute).Return(nil).Once()

		input := InitiateSSOInput{Provider: "malformed"}
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse provider auth url")
		cache.AssertExpectations(t)
	})

	t.Run("failure - cache store state fails", func(t *testing.T) {
		cache := new(mockSSOStateCache)
		uc := NewInitiateSSOUseCase(cache, providers)

		cache.On("StoreState", ctx, mock.Anything, mock.Anything, 10*time.Minute).Return(errors.New("redis error connection refused")).Once()

		input := InitiateSSOInput{Provider: "google"}
		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store sso state: redis error connection refused")
		cache.AssertExpectations(t)
	})
}
