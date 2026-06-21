package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountLockedEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	unlockTime := now.Add(30 * time.Minute)

	event := AccountLockedEvent{
		Email:          "admin@hros.io",
		LockedAt:       now,
		UnlockAt:       unlockTime,
		FailedAttempts: 5,
		IPAddress:      "192.168.1.1",
		UserAgent:      "Mozilla/5.0",
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled AccountLockedEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.True(t, event.LockedAt.Equal(unmarshaled.LockedAt))
	assert.True(t, event.UnlockAt.Equal(unmarshaled.UnlockAt))
	assert.Equal(t, event.FailedAttempts, unmarshaled.FailedAttempts)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
}

func TestEmailSendEvent_Serialization(t *testing.T) {
	event := EmailSendEvent{
		To:       "admin@hros.io",
		Subject:  "Account Locked",
		Template: "account_locked_notification",
		TemplateData: map[string]interface{}{
			"email":     "admin@hros.io",
			"unlock_at": "2026-06-21T23:45:00Z",
		},
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled EmailSendEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.To, unmarshaled.To)
	assert.Equal(t, event.Subject, unmarshaled.Subject)
	assert.Equal(t, event.Template, unmarshaled.Template)
	assert.Equal(t, event.TemplateData["email"], unmarshaled.TemplateData["email"])
	assert.Equal(t, event.TemplateData["unlock_at"], unmarshaled.TemplateData["unlock_at"])
}
