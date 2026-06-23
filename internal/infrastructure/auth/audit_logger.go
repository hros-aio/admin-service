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
func (l *SlogAuditLogger) LogLogoutSuccess(ctx context.Context, _ string) {
	// Do not log the sensitive raw token in logs for security.
	// Only log the event action.
	l.logger.InfoContext(ctx, "logout success",
		slog.String("event", "logout.success"),
	)
}

// LogSessionRefreshed logs a session token refresh event.
func (l *SlogAuditLogger) LogSessionRefreshed(ctx context.Context, userID string) {
	l.logger.InfoContext(ctx, "session refreshed",
		slog.String("event", "session.refreshed"),
		slog.String("user_id", userID),
	)
}

// LogAccountLocked logs an account.locked audit event emitted when brute-force
// protection temporarily locks an email address.
func (l *SlogAuditLogger) LogAccountLocked(ctx context.Context, email string) {
	l.logger.WarnContext(ctx, "account locked",
		slog.String("event", "account.locked"),
		slog.String("email", email),
	)
}

// LogMFAChallengeIssued logs that an MFA challenge has been issued to the user.
func (l *SlogAuditLogger) LogMFAChallengeIssued(ctx context.Context, userID string, email string) {
	l.logger.InfoContext(ctx, "MFA challenge issued",
		slog.String("event", "mfa.challenge_issued"),
		slog.String("user_id", userID),
		slog.String("email", email),
	)
}

// LogMFASuccess logs a successful MFA verification event.
func (l *SlogAuditLogger) LogMFASuccess(ctx context.Context, userID string, email string) {
	l.logger.InfoContext(ctx, "MFA success",
		slog.String("event", "mfa.success"),
		slog.String("user_id", userID),
		slog.String("email", email),
	)
}

// LogMFAFailed logs a failed MFA verification attempt.
func (l *SlogAuditLogger) LogMFAFailed(ctx context.Context, email string, reason string) {
	l.logger.WarnContext(ctx, "MFA failed",
		slog.String("event", "mfa.failed"),
		slog.String("email", email),
		slog.String("reason", reason),
	)
}
