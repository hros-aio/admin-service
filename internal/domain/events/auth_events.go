// Package events defines domain-level event structures.
package events

import (
	"time"
)

// AccountLockedEvent defines the payload structure for the 'account.locked' audit event.
type AccountLockedEvent struct {
	Email          string    `json:"email"`
	LockedAt       time.Time `json:"locked_at"`
	UnlockAt       time.Time `json:"unlock_at"`
	FailedAttempts int       `json:"failed_attempts"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
}

// EmailSendEvent defines the payload structure for the 'email.send' Kafka notification event.
type EmailSendEvent struct {
	To           string                 `json:"to"`
	Subject      string                 `json:"subject"`
	Template     string                 `json:"template"`
	TemplateData map[string]interface{} `json:"template_data"`
}

// MFASuccessEvent defines the payload structure for the 'mfa.success' event.
type MFASuccessEvent struct {
	AdminID    string    `json:"admin_id"`
	Email      string    `json:"email"`
	Method     string    `json:"method"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

// MFAFailedEvent defines the payload structure for the 'mfa.failed' event.
type MFAFailedEvent struct {
	AdminID    string    `json:"admin_id"`
	Email      string    `json:"email"`
	Method     string    `json:"method"`
	Reason     string    `json:"reason"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PasswordResetRequestedEvent defines the payload structure for the 'password.reset_requested' audit event.
type PasswordResetRequestedEvent struct {
	Email      string    `json:"email"`
	Token      string    `json:"token"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PasswordResetCompletedEvent defines the payload structure for the 'password.reset_completed' audit event.
type PasswordResetCompletedEvent struct {
	Email      string    `json:"email"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

// AdminActivatedEvent defines the payload structure for the 'admin.activated' audit event.
type AdminActivatedEvent struct {
	AdminID    string    `json:"admin_id"`
	Email      string    `json:"email"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

// InviteAcceptedEvent defines the payload structure for the 'invite.accepted' audit event.
type InviteAcceptedEvent struct {
	InviteTokenID string    `json:"invite_token_id"`
	AdminID       string    `json:"admin_id"`
	Email         string    `json:"email"`
	InvitedBy     string    `json:"invited_by"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// NotificationSendEvent defines the payload structure for the 'notification.send' Kafka event.
type NotificationSendEvent struct {
	RecipientID string                 `json:"recipient_id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Payload     map[string]interface{} `json:"payload"`
	CreatedAt   time.Time              `json:"created_at"`
}

// SSOSuccessEvent defines the payload structure for the 'login.sso_success' audit event.
type SSOSuccessEvent struct {
	AdminID    string    `json:"admin_id"`
	Email      string    `json:"email"`
	Provider   string    `json:"provider"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

// SSOFailedEvent defines the payload structure for the 'login.sso_failed' audit event.
type SSOFailedEvent struct {
	Email      string    `json:"email,omitempty"`
	Provider   string    `json:"provider"`
	Reason     string    `json:"reason"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

// BiometricSuccessEvent defines the payload structure for the 'login.biometric_success' audit event.
type BiometricSuccessEvent struct {
	AdminID      string    `json:"admin_id"`
	Email        string    `json:"email"`
	CredentialID string    `json:"credential_id"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	OccurredAt   time.Time `json:"occurred_at"`
}
