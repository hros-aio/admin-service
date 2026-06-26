package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain/events"
	"github.com/stretchr/testify/assert"
)

func TestSlogAuditLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	auditLogger := NewSlogAuditLogger(logger)
	ctx := context.Background()

	t.Run("LogLoginSuccess", func(t *testing.T) {
		buf.Reset()
		auditLogger.LogLoginSuccess(ctx, "user-123", "test@example.com")

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "login success", logMap["msg"])
		assert.Equal(t, "login.success", logMap["event"])
		assert.Equal(t, "user-123", logMap["user_id"])
		assert.Equal(t, "test@example.com", logMap["email"])
	})

	t.Run("LogLoginFailed", func(t *testing.T) {
		buf.Reset()
		auditLogger.LogLoginFailed(ctx, "test@example.com", "invalid password")

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "login failed", logMap["msg"])
		assert.Equal(t, "login.failed", logMap["event"])
		assert.Equal(t, "test@example.com", logMap["email"])
		assert.Equal(t, "invalid password", logMap["reason"])
	})

	t.Run("LogLogoutSuccess", func(t *testing.T) {
		buf.Reset()
		auditLogger.LogLogoutSuccess(ctx, "some-refresh-token")

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "logout success", logMap["msg"])
		assert.Equal(t, "logout.success", logMap["event"])
		// Verify that the sensitive token is NOT logged in plain text
		for _, v := range logMap {
			assert.NotEqual(t, "some-refresh-token", v)
		}
	})

	t.Run("LogSessionRefreshed", func(t *testing.T) {
		buf.Reset()
		auditLogger.LogSessionRefreshed(ctx, "user-123")

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "session refreshed", logMap["msg"])
		assert.Equal(t, "session.refreshed", logMap["event"])
		assert.Equal(t, "user-123", logMap["user_id"])
	})

	t.Run("LogAccountLocked", func(t *testing.T) {
		buf.Reset()
		auditLogger.LogAccountLocked(ctx, "locked@example.com")

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "account locked", logMap["msg"])
		assert.Equal(t, "account.locked", logMap["event"])
		assert.Equal(t, "locked@example.com", logMap["email"])
	})

	t.Run("LogMFAChallengeIssued", func(t *testing.T) {
		buf.Reset()
		auditLogger.LogMFAChallengeIssued(ctx, "user-123", "test@example.com")

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "MFA challenge issued", logMap["msg"])
		assert.Equal(t, "mfa.challenge_issued", logMap["event"])
		assert.Equal(t, "user-123", logMap["user_id"])
		assert.Equal(t, "test@example.com", logMap["email"])
	})

	t.Run("LogMFASuccess", func(t *testing.T) {
		buf.Reset()
		auditLogger.LogMFASuccess(ctx, "user-123", "test@example.com")

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "MFA success", logMap["msg"])
		assert.Equal(t, "mfa.success", logMap["event"])
		assert.Equal(t, "user-123", logMap["user_id"])
		assert.Equal(t, "test@example.com", logMap["email"])
	})

	t.Run("LogMFAFailed", func(t *testing.T) {
		buf.Reset()
		auditLogger.LogMFAFailed(ctx, "test@example.com", "invalid code")

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "MFA failed", logMap["msg"])
		assert.Equal(t, "mfa.failed", logMap["event"])
		assert.Equal(t, "test@example.com", logMap["email"])
		assert.Equal(t, "invalid code", logMap["reason"])
	})

	t.Run("LogPasswordResetRequested", func(t *testing.T) {
		buf.Reset()
		event := events.PasswordResetRequestedEvent{
			Email:      "reset@example.com",
			Token:      "secure-token",
			IPAddress:  "127.0.0.1",
			UserAgent:  "Mozilla/5.0",
			OccurredAt: time.Now(),
		}
		auditLogger.LogPasswordResetRequested(ctx, event)

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "password reset requested", logMap["msg"])
		assert.Equal(t, "password.reset_requested", logMap["event"])
		assert.Equal(t, "reset@example.com", logMap["email"])
		assert.Equal(t, "127.0.0.1", logMap["ip_address"])
		assert.Equal(t, "Mozilla/5.0", logMap["user_agent"])
		assert.Equal(t, "[REDACTED]", logMap["token"])
	})

	t.Run("LogPasswordResetCompleted", func(t *testing.T) {
		buf.Reset()
		event := events.PasswordResetCompletedEvent{
			Email:      "complete@example.com",
			IPAddress:  "10.0.0.1",
			UserAgent:  "Firefox",
			OccurredAt: time.Now(),
		}
		auditLogger.LogPasswordResetCompleted(ctx, event)

		var logMap map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logMap)
		assert.NoError(t, err)

		assert.Equal(t, "password reset completed", logMap["msg"])
		assert.Equal(t, "password.reset_completed", logMap["event"])
		assert.Equal(t, "complete@example.com", logMap["email"])
		assert.Equal(t, "10.0.0.1", logMap["ip_address"])
		assert.Equal(t, "Firefox", logMap["user_agent"])
	})
}
