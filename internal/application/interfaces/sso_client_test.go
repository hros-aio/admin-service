package interfaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSOUserProfile(t *testing.T) {
	profile := &SSOUserProfile{
		Email:      "user@example.com",
		IdentityID: "subject-123",
		Provider:   "google",
	}

	assert.Equal(t, "user@example.com", profile.Email)
	assert.Equal(t, "subject-123", profile.IdentityID)
	assert.Equal(t, "google", profile.Provider)
}
