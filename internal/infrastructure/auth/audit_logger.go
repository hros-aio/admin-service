package auth

import (
	"context"
	"log/slog"

	"github.com/hros/admin-service/internal/domain/auth"
	"github.com/hros/admin-service/internal/domain/events"
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

// LogPasswordResetRequested logs a password reset request event.
func (l *SlogAuditLogger) LogPasswordResetRequested(ctx context.Context, event events.PasswordResetRequestedEvent) {
	l.logger.InfoContext(ctx, "password reset requested",
		slog.String("event", "password.reset_requested"),
		slog.String("email", event.Email),
		slog.String("ip_address", event.IPAddress),
		slog.String("user_agent", event.UserAgent),
		slog.String("token", "[REDACTED]"), // Redact token for security
	)
}

// LogPasswordResetCompleted logs a completed password reset event.
func (l *SlogAuditLogger) LogPasswordResetCompleted(ctx context.Context, event events.PasswordResetCompletedEvent) {
	l.logger.InfoContext(ctx, "password reset completed",
		slog.String("event", "password.reset_completed"),
		slog.String("email", event.Email),
		slog.String("ip_address", event.IPAddress),
		slog.String("user_agent", event.UserAgent),
	)
}

// LogInviteAccepted logs that an administrator accepted an invitation.
func (l *SlogAuditLogger) LogInviteAccepted(ctx context.Context, event events.InviteAcceptedEvent) {
	l.logger.InfoContext(ctx, "invite accepted",
		slog.String("event", "invite.accepted"),
		slog.String("invite_token_id", event.InviteTokenID),
		slog.String("admin_id", event.AdminID),
		slog.String("email", event.Email),
		slog.String("invited_by", event.InvitedBy),
		slog.String("ip_address", event.IPAddress),
		slog.String("user_agent", event.UserAgent),
	)
}

// LogAdminActivated logs that an administrator account was successfully activated.
func (l *SlogAuditLogger) LogAdminActivated(ctx context.Context, event events.AdminActivatedEvent) {
	l.logger.InfoContext(ctx, "admin account activated",
		slog.String("event", "admin.activated"),
		slog.String("admin_id", event.AdminID),
		slog.String("email", event.Email),
		slog.String("ip_address", event.IPAddress),
		slog.String("user_agent", event.UserAgent),
	)
}

// LogSSOSuccess logs a successful SSO login event.
func (l *SlogAuditLogger) LogSSOSuccess(ctx context.Context, event events.SSOSuccessEvent) {
	l.logger.InfoContext(ctx, "SSO success",
		slog.String("event", "login.sso_success"),
		slog.String("admin_id", event.AdminID),
		slog.String("email", event.Email),
		slog.String("provider", event.Provider),
		slog.String("ip_address", event.IPAddress),
		slog.String("user_agent", event.UserAgent),
		slog.Time("occurred_at", event.OccurredAt),
	)
}

// LogSSOFailed logs a failed SSO login event.
func (l *SlogAuditLogger) LogSSOFailed(ctx context.Context, event events.SSOFailedEvent) {
	l.logger.WarnContext(ctx, "SSO failed",
		slog.String("event", "login.sso_failed"),
		slog.String("email", event.Email),
		slog.String("provider", event.Provider),
		slog.String("reason", event.Reason),
		slog.String("ip_address", event.IPAddress),
		slog.String("user_agent", event.UserAgent),
		slog.Time("occurred_at", event.OccurredAt),
	)
}

// LogBiometricSuccess logs a successful biometric login event.
func (l *SlogAuditLogger) LogBiometricSuccess(ctx context.Context, event events.BiometricSuccessEvent) {
	l.logger.InfoContext(ctx, "biometric login success",
		slog.String("event", "login.biometric_success"),
		slog.String("admin_id", event.AdminID),
		slog.String("email", event.Email),
		slog.String("credential_id", event.CredentialID),
		slog.String("ip_address", event.IPAddress),
		slog.String("user_agent", event.UserAgent),
		slog.Time("occurred_at", event.OccurredAt),
	)
}
