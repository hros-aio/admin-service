package producer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/IBM/sarama/mocks"
	"github.com/hros/admin-service/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestLogger returns a no-op logger suitable for unit tests.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newValidEvent returns a fully populated EmailSendEvent for test use.
func newValidEvent() events.EmailSendEvent {
	return events.EmailSendEvent{
		To:       "admin@hros.io",
		Subject:  "Account Locked",
		Template: "account_locked_notification",
		TemplateData: map[string]interface{}{
			"email":     "admin@hros.io",
			"unlock_at": "2026-06-21T17:15:00Z",
		},
	}
}

func TestEmailKafkaProducer_PublishLockoutEmail_HappyPath(t *testing.T) {
	mock := mocks.NewSyncProducer(t, nil)
	// Expect exactly one message to be sent successfully.
	mock.ExpectSendMessageAndSucceed()

	producer := NewEmailKafkaProducer(mock, newTestLogger())

	event := newValidEvent()
	err := producer.PublishLockoutEmail(context.Background(), event)
	require.NoError(t, err)

	// Verify mock expectations are met (called exactly once).
	require.NoError(t, mock.Close())
}

func TestEmailKafkaProducer_PublishLockoutEmail_EnvelopeShape(t *testing.T) {
	// Capture the ProducerMessage to inspect the serialized envelope.
	var capturedMsg []byte

	mock := mocks.NewSyncProducer(t, nil)
	mock.ExpectSendMessageWithCheckerFunctionAndSucceed(func(msg []byte) error {
		capturedMsg = msg
		return nil
	})

	producer := NewEmailKafkaProducer(mock, newTestLogger())

	event := newValidEvent()
	err := producer.PublishLockoutEmail(context.Background(), event)
	require.NoError(t, err)
	require.NoError(t, mock.Close())

	// Unmarshal and inspect envelope.
	var envelope EventEnvelope[events.EmailSendEvent]
	err = json.Unmarshal(capturedMsg, &envelope)
	require.NoError(t, err, "envelope must deserialize cleanly")

	assert.NotEmpty(t, envelope.ID, "ID must be a non-empty UUID")
	assert.Equal(t, emailSendEventType, envelope.Type, "Type must be 'email.send'")
	assert.Equal(t, defaultSource, envelope.Source, "Source must be 'admin-service'")
	assert.Equal(t, envelopeVersion, envelope.Version, "Version must be 1")
	assert.False(t, envelope.OccurredAt.IsZero(), "OccurredAt must be non-zero")

	// Verify the domain payload is preserved exactly.
	assert.Equal(t, event.To, envelope.Data.To)
	assert.Equal(t, event.Subject, envelope.Data.Subject)
	assert.Equal(t, event.Template, envelope.Data.Template)
	assert.Equal(t, event.TemplateData["email"], envelope.Data.TemplateData["email"])
	assert.Equal(t, event.TemplateData["unlock_at"], envelope.Data.TemplateData["unlock_at"])
}

func TestEmailKafkaProducer_PublishLockoutEmail_SaramaError(t *testing.T) {
	brokerErr := errors.New("broker unavailable")

	mock := mocks.NewSyncProducer(t, nil)
	mock.ExpectSendMessageAndFail(brokerErr)

	producer := NewEmailKafkaProducer(mock, newTestLogger())

	err := producer.PublishLockoutEmail(context.Background(), newValidEvent())
	require.Error(t, err)
	assert.ErrorContains(t, err, "publish lockout email")
	assert.ErrorContains(t, err, "broker unavailable")

	require.NoError(t, mock.Close())
}

func TestEmailKafkaProducer_PublishLockoutEmail_EmptyRecipient(t *testing.T) {
	// MockSyncProducer is created without any expectations.
	// If SendMessage is called unexpectedly, the mock will fail the test.
	mock := mocks.NewSyncProducer(t, nil)

	producer := NewEmailKafkaProducer(mock, newTestLogger())

	event := newValidEvent()
	event.To = ""

	err := producer.PublishLockoutEmail(context.Background(), event)
	require.Error(t, err)
	assert.ErrorContains(t, err, "recipient email must not be empty")

	// Close with no consumed expectations — mock asserts nothing unexpected was called.
	require.NoError(t, mock.Close())
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"normal email", "admin@hros.io"},
		{"empty string", ""},
		{"different emails produce different masks", "other@hros.io"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskEmail(tt.email)
			if tt.email == "" {
				assert.Empty(t, result)
				return
			}
			assert.Len(t, result, 12, "mask must be exactly 12 hex chars")
			// Mask must not contain the original email.
			assert.NotContains(t, result, "hros.io")
			assert.NotContains(t, result, "admin")
		})
	}
}

func TestMaskEmail_Deterministic(t *testing.T) {
	email := "admin@hros.io"
	// Same input must always produce same output (no random salt).
	assert.Equal(t, maskEmail(email), maskEmail(email))
	// Different input must produce different output.
	assert.NotEqual(t, maskEmail("admin@hros.io"), maskEmail("other@hros.io"))
}
