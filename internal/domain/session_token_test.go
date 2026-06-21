package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionToken_IsExpired(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"not expired", future, false},
		{"expired", past, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &SessionToken{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.want, token.IsExpired())
		})
	}
}

func TestSessionToken_IsRevoked(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		revokedAt *time.Time
		want      bool
	}{
		{"not revoked", nil, false},
		{"revoked", &now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &SessionToken{RevokedAt: tt.revokedAt}
			assert.Equal(t, tt.want, token.IsRevoked())
		})
	}
}

type errorReader struct{}

func (errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("forced rand failure")
}

func TestSessionToken_Rotate(t *testing.T) {
	now := time.Now()
	newExpiry := now.Add(2 * time.Hour)

	token := &SessionToken{
		RefreshToken: "old-token",
		ExpiresAt:    now,
	}

	newToken, err := token.Rotate(newExpiry)
	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.Equal(t, newToken, token.RefreshToken)
	assert.Equal(t, newExpiry, token.ExpiresAt)

	// Refresh token must be 64-character hex string (representing 32 random bytes)
	assert.Len(t, newToken, 64)
	_, err = hex.DecodeString(newToken)
	assert.NoError(t, err)
}

func TestSessionToken_Rotate_RandFailure(t *testing.T) {
	oldReader := rand.Reader
	defer func() { rand.Reader = oldReader }()
	rand.Reader = errorReader{}

	expiry := time.Now()
	token := &SessionToken{
		RefreshToken: "old-token",
		ExpiresAt:    expiry,
	}

	_, err := token.Rotate(expiry.Add(time.Hour))
	assert.Error(t, err)
	assert.Equal(t, "old-token", token.RefreshToken) // Should not mutate state on failure
	assert.Equal(t, expiry, token.ExpiresAt)         // Should not mutate expiration on failure
}
