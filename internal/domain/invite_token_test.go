package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInviteToken_IsExpired(t *testing.T) {
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
			token := &InviteToken{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.want, token.IsExpired())
		})
	}
}

func TestInviteToken_IsUsed(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		usedAt *time.Time
		want   bool
	}{
		{"not used", nil, false},
		{"used", &now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &InviteToken{UsedAt: tt.usedAt}
			assert.Equal(t, tt.want, token.IsUsed())
		})
	}
}

func TestInviteToken_Consume(t *testing.T) {
	token := &InviteToken{
		UsedAt: nil,
	}

	assert.False(t, token.IsUsed())
	token.Consume()
	assert.True(t, token.IsUsed())
	assert.NotNil(t, token.UsedAt)
	assert.WithinDuration(t, time.Now(), *token.UsedAt, time.Second)
}
