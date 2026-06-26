package auth

import (
	"context"

	"github.com/hros/admin-service/internal/domain/events"
)

// AuditLogger defines the interface for logging security-relevant events.
type AuditLogger interface {
	LogLoginSuccess(ctx context.Context, userID string, email string)
	LogLoginFailed(ctx context.Context, email string, reason string)
	LogLogoutSuccess(ctx context.Context, token string)
	LogSessionRefreshed(ctx context.Context, userID string)
	// LogAccountLocked records that an account was temporarily locked due to brute-force protection.
	LogAccountLocked(ctx context.Context, email string)
	// LogMFAChallengeIssued records that an MFA challenge was successfully generated and issued.
	LogMFAChallengeIssued(ctx context.Context, userID string, email string)
	// LogMFASuccess records that a user successfully completed the second factor authentication.
	LogMFASuccess(ctx context.Context, userID string, email string)
	// LogMFAFailed records that a user failed the second factor authentication.
	LogMFAFailed(ctx context.Context, email string, reason string)
	// LogPasswordResetRequested records a password reset request.
	LogPasswordResetRequested(ctx context.Context, event events.PasswordResetRequestedEvent)
}
