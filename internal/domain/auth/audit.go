package auth

import "context"

// AuditLogger defines the interface for logging security-relevant events.
type AuditLogger interface {
	LogLoginSuccess(ctx context.Context, userID string, email string)
	LogLoginFailed(ctx context.Context, email string, reason string)
	LogLogoutSuccess(ctx context.Context, token string)
	LogSessionRefreshed(ctx context.Context, userID string)
	// LogAccountLocked records that an account was temporarily locked due to brute-force protection.
	LogAccountLocked(ctx context.Context, email string)
}
