package auth

import (
	"testing"

	"github.com/hros/admin-service/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestProvideSSOProviders(t *testing.T) {
	t.Run("all empty", func(t *testing.T) {
		cfg := &config.Config{}
		providers, err := ProvideSSOProviders(cfg)
		assert.NoError(t, err)
		assert.Empty(t, providers)
	})

	t.Run("all set", func(t *testing.T) {
		cfg := &config.Config{
			SSOGoogleClientID:    "client-id",
			SSOGoogleRedirectURL: "redirect-url",
			SSOGoogleAuthURL:     "auth-url",
		}
		providers, err := ProvideSSOProviders(cfg)
		assert.NoError(t, err)
		assert.Len(t, providers, 1)
		google, exists := providers["google"]
		assert.True(t, exists)
		assert.Equal(t, "client-id", google.ClientID)
		assert.Equal(t, "redirect-url", google.RedirectURL)
		assert.Equal(t, "auth-url", google.AuthURL)
	})

	t.Run("partial - client id missing", func(t *testing.T) {
		cfg := &config.Config{
			SSOGoogleRedirectURL: "redirect-url",
			SSOGoogleAuthURL:     "auth-url",
		}
		providers, err := ProvideSSOProviders(cfg)
		assert.Error(t, err)
		assert.Nil(t, providers)
		assert.Contains(t, err.Error(), "incomplete")
	})

	t.Run("partial - redirect url missing", func(t *testing.T) {
		cfg := &config.Config{
			SSOGoogleClientID: "client-id",
			SSOGoogleAuthURL:  "auth-url",
		}
		providers, err := ProvideSSOProviders(cfg)
		assert.Error(t, err)
		assert.Nil(t, providers)
		assert.Contains(t, err.Error(), "incomplete")
	})

	t.Run("partial - auth url missing", func(t *testing.T) {
		cfg := &config.Config{
			SSOGoogleClientID:    "client-id",
			SSOGoogleRedirectURL: "redirect-url",
		}
		providers, err := ProvideSSOProviders(cfg)
		assert.Error(t, err)
		assert.Nil(t, providers)
		assert.Contains(t, err.Error(), "incomplete")
	})
}
