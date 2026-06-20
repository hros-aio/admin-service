// Package auth implements security and authentication infrastructure adapters.
package auth

import (
	"context"
	"log/slog"

	"github.com/hros/admin-service/internal/domain/auth"
)

// SlogAuditLogger implements the AuditLogger interface using slog.
type SlogAuditLogger struct {
	logger *slog.Logger
}

// NewSlogAuditLogger creates a new SlogAuditLogger.
func NewSlogAuditLogger(logger *slog.Logger) auth.AuditLogger {
	return &SlogAuditLogger{logger: logger}
}

// LogLoginSuccess logs a successful login event.
func (l *SlogAuditLogger) LogLoginSuccess(ctx context.Context, userID string, email string) {
	l.logger.InfoContext(ctx, "login success",
		slog.String("event", "login.success"),
		slog.String("user_id", userID),
		slog.String("email", email),
	)
}

// LogLoginFailed logs a failed login attempt.
func (l *SlogAuditLogger) LogLoginFailed(ctx context.Context, email string, reason string) {
	l.logger.WarnContext(ctx, "login failed",
		slog.String("event", "login.failed"),
		slog.String("email", email),
		slog.String("reason", reason),
	)
}

// LogLogoutSuccess logs a successful logout event.
func (l *SlogAuditLogger) LogLogoutSuccess(ctx context.Context, token string) {
	// Mask the token or use a hash for safety. We never log full raw secrets.
	maskedToken := "unknown"
	if len(token) > 8 {
		maskedToken = token[:8] + "..."
	}
	l.logger.InfoContext(ctx, "logout success",
		slog.String("event", "logout.success"),
		slog.String("token_prefix", maskedToken),
	)
}
