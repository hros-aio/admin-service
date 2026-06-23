package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdminUser_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status AdminUserStatus
		want   bool
	}{
		{"active", AdminUserStatusActive, true},
		{"inactive", AdminUserStatusInactive, false},
		{"pending", AdminUserStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &AdminUser{Status: tt.status}
			assert.Equal(t, tt.want, u.IsActive())
		})
	}
}

func TestAdminUser_IsLocked(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name        string
		lockedUntil *time.Time
		want        bool
	}{
		{"not locked", nil, false},
		{"locked in future", &future, true},
		{"locked in past", &past, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &AdminUser{LockedUntil: tt.lockedUntil}
			assert.Equal(t, tt.want, u.IsLocked())
		})
	}
}

func TestAdminUser_MFAFields(t *testing.T) {
	u := &AdminUser{
		TotpSecret:          "secret_totp_key",
		WebauthnCredentials: []byte(`{"id":"cred_123"}`),
	}

	assert.Equal(t, "secret_totp_key", u.TotpSecret)
	assert.Equal(t, []byte(`{"id":"cred_123"}`), u.WebauthnCredentials)
}
