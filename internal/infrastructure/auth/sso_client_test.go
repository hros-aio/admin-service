package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSSOClient_ExchangeCode(t *testing.T) {
	client := NewDefaultSSOClient()
	profile, err := client.ExchangeCode(context.Background(), "google", "code")

	assert.Nil(t, profile)
	assert.Error(t, err)
	assert.Equal(t, "sso client code exchange not implemented", err.Error())
}
