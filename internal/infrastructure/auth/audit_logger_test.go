package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

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
}
