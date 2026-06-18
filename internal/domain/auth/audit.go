package auth

import "context"

// AuditLogger defines the interface for logging security-relevant events.
type AuditLogger interface {
	LogLoginSuccess(ctx context.Context, userID string, email string)
	LogLoginFailed(ctx context.Context, email string, reason string)
}
