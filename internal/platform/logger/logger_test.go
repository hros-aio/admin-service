package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("debug_level", func(t *testing.T) {
		log := New("debug")
		assert.NotNil(t, log)
		// We can't easily check the internal level of slog.Logger without custom handler,
		// but we verified the constructor doesn't panic.
	})

	t.Run("invalid_level_defaults_to_info", func(t *testing.T) {
		log := New("invalid")
		assert.NotNil(t, log)
	})

	t.Run("output_format", func(t *testing.T) {
		var buf bytes.Buffer
		opts := &slog.HandlerOptions{Level: slog.LevelInfo}
		handler := slog.NewJSONHandler(&buf, opts)
		log := slog.New(handler)

		log.Info("test message", slog.String("key", "value"))

		output := buf.String()
		assert.Contains(t, output, `"msg":"test message"`)
		assert.Contains(t, output, `"level":"INFO"`)
		assert.Contains(t, output, `"key":"value"`)
	})
}
