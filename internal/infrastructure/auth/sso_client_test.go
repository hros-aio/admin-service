package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSSOClient_ExchangeCode(t *testing.T) {
	client := NewDefaultSSOClient()

	t.Run("success exchange", func(t *testing.T) {
		profile, err := client.ExchangeCode(context.Background(), "google", "code-123")
		assert.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "sso-user@example.com", profile.Email)
		assert.Equal(t, "sso-identity-123", profile.IdentityID)
		assert.Equal(t, "google", profile.Provider)
	})

	t.Run("empty provider", func(t *testing.T) {
		profile, err := client.ExchangeCode(context.Background(), "", "code-123")
		assert.Error(t, err)
		assert.Nil(t, profile)
	})

	t.Run("empty code", func(t *testing.T) {
		profile, err := client.ExchangeCode(context.Background(), "google", "")
		assert.Error(t, err)
		assert.Nil(t, profile)
	})
}
