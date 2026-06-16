package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
