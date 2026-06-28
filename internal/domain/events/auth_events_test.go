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

func TestMFASuccessEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := MFASuccessEvent{
		AdminID:    "admin_123",
		Email:      "admin@hros.io",
		Method:     "totp",
		IPAddress:  "192.168.1.10",
		UserAgent:  "Firefox",
		OccurredAt: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled MFASuccessEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.AdminID, unmarshaled.AdminID)
	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.Method, unmarshaled.Method)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

func TestMFAFailedEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := MFAFailedEvent{
		AdminID:    "admin_123",
		Email:      "admin@hros.io",
		Method:     "webauthn",
		Reason:     "Invalid signature",
		IPAddress:  "192.168.1.10",
		UserAgent:  "Safari",
		OccurredAt: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled MFAFailedEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.AdminID, unmarshaled.AdminID)
	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.Method, unmarshaled.Method)
	assert.Equal(t, event.Reason, unmarshaled.Reason)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

func TestPasswordResetRequestedEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := PasswordResetRequestedEvent{
		Email:      "admin@hros.io",
		Token:      "reset_token_123",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		OccurredAt: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled PasswordResetRequestedEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.Token, unmarshaled.Token)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

func TestPasswordResetCompletedEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := PasswordResetCompletedEvent{
		Email:      "admin@hros.io",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		OccurredAt: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled PasswordResetCompletedEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

func TestAdminActivatedEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := AdminActivatedEvent{
		AdminID:    "admin-uuid-123",
		Email:      "admin@hros.io",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		OccurredAt: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled AdminActivatedEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.AdminID, unmarshaled.AdminID)
	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

func TestInviteAcceptedEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := InviteAcceptedEvent{
		InviteTokenID: "token-uuid-456",
		AdminID:       "admin-uuid-123",
		Email:         "admin@hros.io",
		InvitedBy:     "superadmin-uuid-789",
		IPAddress:     "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
		OccurredAt:    now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled InviteAcceptedEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.InviteTokenID, unmarshaled.InviteTokenID)
	assert.Equal(t, event.AdminID, unmarshaled.AdminID)
	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.InvitedBy, unmarshaled.InvitedBy)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

func TestNotificationSendEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := NotificationSendEvent{
		RecipientID: "superadmin-uuid-789",
		Type:        "invite_accepted",
		Title:       "Invitation Accepted",
		Message:     "admin@hros.io has accepted your invitation.",
		Payload: map[string]interface{}{
			"admin_id": "admin-uuid-123",
		},
		CreatedAt: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled NotificationSendEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.RecipientID, unmarshaled.RecipientID)
	assert.Equal(t, event.Type, unmarshaled.Type)
	assert.Equal(t, event.Title, unmarshaled.Title)
	assert.Equal(t, event.Message, unmarshaled.Message)
	assert.Equal(t, event.Payload["admin_id"], unmarshaled.Payload["admin_id"])
	assert.True(t, event.CreatedAt.Equal(unmarshaled.CreatedAt))
}

func TestSSOSuccessEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := SSOSuccessEvent{
		AdminID:    "admin-uuid-123",
		Email:      "admin@hros.io",
		Provider:   "okta",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		OccurredAt: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled SSOSuccessEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.AdminID, unmarshaled.AdminID)
	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.Provider, unmarshaled.Provider)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

func TestSSOFailedEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := SSOFailedEvent{
		Email:      "admin@hros.io",
		Provider:   "google",
		Reason:     "no_account_linked",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		OccurredAt: now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled SSOFailedEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.Provider, unmarshaled.Provider)
	assert.Equal(t, event.Reason, unmarshaled.Reason)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

func TestBiometricSuccessEvent_Serialization(t *testing.T) {
	now := time.Now().UTC()
	event := BiometricSuccessEvent{
		AdminID:      "admin-uuid-123",
		Email:        "admin@hros.io",
		CredentialID: "cred-id-abc",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		OccurredAt:   now,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled BiometricSuccessEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.AdminID, unmarshaled.AdminID)
	assert.Equal(t, event.Email, unmarshaled.Email)
	assert.Equal(t, event.CredentialID, unmarshaled.CredentialID)
	assert.Equal(t, event.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, event.UserAgent, unmarshaled.UserAgent)
	assert.True(t, event.OccurredAt.Equal(unmarshaled.OccurredAt))
}

