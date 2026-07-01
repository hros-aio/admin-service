// Package producer implements Kafka event producer adapters for the admin service.
package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/hros/admin-service/internal/domain/events"
)

const (
	// notificationSendTopic is the Kafka topic for in-app notification events.
	notificationSendTopic = "notification.send.v1"

	// notificationSendEventType is the event type string embedded in the envelope.
	notificationSendEventType = "notification.send"
)

// NotificationKafkaProducer adapts domain NotificationSendEvent values into Kafka
// messages using the standard EventEnvelope and dispatches them via a SyncProducer.
type NotificationKafkaProducer struct {
	producer sarama.SyncProducer
	source   string
	logger   *slog.Logger
}

// NewNotificationKafkaProducer creates a NotificationKafkaProducer.
func NewNotificationKafkaProducer(producer sarama.SyncProducer, logger *slog.Logger) *NotificationKafkaProducer {
	return &NotificationKafkaProducer{
		producer: producer,
		source:   defaultSource,
		logger:   logger,
	}
}

// PublishInviteAcceptedNotification wraps a NotificationSendEvent in a standard
// EventEnvelope and publishes it to the notification.send.v1 Kafka topic.
// The message key is the RecipientID to guarantee per-recipient ordering.
//
// Returns a validation error immediately if event.RecipientID is empty.
// Returns a wrapped error if Sarama fails to deliver the message.
// On success, logs the published event at Info level.
func (p *NotificationKafkaProducer) PublishInviteAcceptedNotification(ctx context.Context, event events.NotificationSendEvent) error {
	if event.RecipientID == "" {
		return fmt.Errorf("recipient ID must not be empty")
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	envelope := EventEnvelope[events.NotificationSendEvent]{
		ID:            newUUID(),
		Type:          notificationSendEventType,
		Source:        p.source,
		Version:       envelopeVersion,
		CorrelationID: correlationIDFromContext(ctx),
		OccurredAt:    time.Now().UTC(),
		Data:          event,
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal invite accepted notification envelope: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: notificationSendTopic,
		Key:   sarama.StringEncoder(event.RecipientID),
		Value: sarama.ByteEncoder(payload),
	}

	if p.producer == nil {
		p.logger.InfoContext(
			ctx, "Kafka producer is disabled (KAFKA_PRODUCE_ENABLE is not true), skipping publish of invite accepted notification event",
			slog.String("event", "kafka.notification_send.skipped"),
			slog.String("topic", notificationSendTopic),
			slog.String("recipient_id", event.RecipientID),
			slog.String("type", event.Type),
		)
		return nil
	}

	if _, _, err := p.producer.SendMessage(msg); err != nil {
		return fmt.Errorf("publish invite accepted notification: %w", err)
	}

	p.logger.InfoContext(
		ctx, "invite accepted notification event published to Kafka",
		slog.String("event", "kafka.notification_send.published"),
		slog.String("topic", notificationSendTopic),
		slog.String("recipient_id", event.RecipientID),
		slog.String("type", event.Type),
	)

	return nil
}
