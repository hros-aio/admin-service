package producer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama/mocks"
	"github.com/hros/admin-service/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newValidNotificationEvent returns a fully populated NotificationSendEvent for test use.
func newValidNotificationEvent() events.NotificationSendEvent {
	return events.NotificationSendEvent{
		RecipientID: "inviter-uuid-001",
		Type:        "invite.accepted",
		Title:       "Your invite was accepted",
		Message:     "Admin user admin@hros.io has accepted your invitation and activated their account.",
		Payload: map[string]interface{}{
			"admin_id": "new-admin-uuid",
			"email":    "admin@hros.io",
		},
		CreatedAt: time.Now().UTC(),
	}
}

func TestNotificationKafkaProducer_PublishInviteAcceptedNotification_HappyPath(t *testing.T) {
	mock := mocks.NewSyncProducer(t, nil)
	mock.ExpectSendMessageAndSucceed()

	producer := NewNotificationKafkaProducer(mock, newTestLogger())

	event := newValidNotificationEvent()
	err := producer.PublishInviteAcceptedNotification(context.Background(), event)
	require.NoError(t, err)

	require.NoError(t, mock.Close())
}

func TestNotificationKafkaProducer_PublishInviteAcceptedNotification_EnvelopeShape(t *testing.T) {
	var capturedMsg []byte

	mock := mocks.NewSyncProducer(t, nil)
	mock.ExpectSendMessageWithCheckerFunctionAndSucceed(func(msg []byte) error {
		capturedMsg = msg
		return nil
	})

	producer := NewNotificationKafkaProducer(mock, newTestLogger())

	event := newValidNotificationEvent()
	err := producer.PublishInviteAcceptedNotification(context.Background(), event)
	require.NoError(t, err)
	require.NoError(t, mock.Close())

	// Unmarshal and inspect envelope fields.
	var envelope EventEnvelope[events.NotificationSendEvent]
	err = json.Unmarshal(capturedMsg, &envelope)
	require.NoError(t, err, "envelope must deserialize cleanly")

	assert.NotEmpty(t, envelope.ID, "ID must be a non-empty UUID")
	assert.Equal(t, notificationSendEventType, envelope.Type, "Type must be 'notification.send'")
	assert.Equal(t, defaultSource, envelope.Source, "Source must be 'admin-service'")
	assert.Equal(t, envelopeVersion, envelope.Version, "Version must be 1")
	assert.False(t, envelope.OccurredAt.IsZero(), "OccurredAt must be non-zero")

	// Verify the domain payload is preserved exactly.
	assert.Equal(t, event.RecipientID, envelope.Data.RecipientID)
	assert.Equal(t, event.Type, envelope.Data.Type)
	assert.Equal(t, event.Title, envelope.Data.Title)
	assert.Equal(t, event.Message, envelope.Data.Message)
	assert.Equal(t, event.Payload["admin_id"], envelope.Data.Payload["admin_id"])
	assert.Equal(t, event.Payload["email"], envelope.Data.Payload["email"])
}

func TestNotificationKafkaProducer_PublishInviteAcceptedNotification_SaramaError(t *testing.T) {
	brokerErr := errors.New("broker unavailable")

	mock := mocks.NewSyncProducer(t, nil)
	mock.ExpectSendMessageAndFail(brokerErr)

	producer := NewNotificationKafkaProducer(mock, newTestLogger())

	err := producer.PublishInviteAcceptedNotification(context.Background(), newValidNotificationEvent())
	require.Error(t, err)
	assert.ErrorContains(t, err, "publish invite accepted notification")
	assert.ErrorContains(t, err, "broker unavailable")

	require.NoError(t, mock.Close())
}

func TestNotificationKafkaProducer_PublishInviteAcceptedNotification_EmptyRecipient(t *testing.T) {
	// No expectations set — if SendMessage is called, the mock will fail the test.
	mock := mocks.NewSyncProducer(t, nil)

	producer := NewNotificationKafkaProducer(mock, newTestLogger())

	event := newValidNotificationEvent()
	event.RecipientID = ""

	err := producer.PublishInviteAcceptedNotification(context.Background(), event)
	require.Error(t, err)
	assert.ErrorContains(t, err, "recipient ID must not be empty")

	require.NoError(t, mock.Close())
}

func TestNotificationKafkaProducer_PublishInviteAcceptedNotification_MarshalError(t *testing.T) {
	mock := mocks.NewSyncProducer(t, nil)
	producer := NewNotificationKafkaProducer(mock, newTestLogger())

	event := newValidNotificationEvent()
	// Channel values cannot be serialized to JSON, causing a marshal error.
	event.Payload = map[string]interface{}{
		"bad_field": make(chan int),
	}

	err := producer.PublishInviteAcceptedNotification(context.Background(), event)
	require.Error(t, err)
	assert.ErrorContains(t, err, "marshal invite accepted notification envelope")

	require.NoError(t, mock.Close())
}

func TestNotificationKafkaProducer_PublishInviteAcceptedNotification_ZeroCreatedAt(t *testing.T) {
	var capturedMsg []byte

	mock := mocks.NewSyncProducer(t, nil)
	mock.ExpectSendMessageWithCheckerFunctionAndSucceed(func(msg []byte) error {
		capturedMsg = msg
		return nil
	})

	producer := NewNotificationKafkaProducer(mock, newTestLogger())

	// CreatedAt is zero — producer should auto-fill it.
	event := newValidNotificationEvent()
	event.CreatedAt = time.Time{}

	err := producer.PublishInviteAcceptedNotification(context.Background(), event)
	require.NoError(t, err)
	require.NoError(t, mock.Close())

	var envelope EventEnvelope[events.NotificationSendEvent]
	err = json.Unmarshal(capturedMsg, &envelope)
	require.NoError(t, err)

	assert.False(t, envelope.Data.CreatedAt.IsZero(), "CreatedAt must be set automatically when zero")
}
