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
