// Package producer implements Kafka event producer adapters for the admin service.
package producer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/hros/admin-service/internal/domain/events"
)

const (
	// emailSendTopic is the Kafka topic for lockout email notification events.
	emailSendTopic = "email.send.v1"

	// emailSendEventType is the event type string embedded in the envelope.
	emailSendEventType = "email.send"

	// defaultSource identifies the publishing service in every event envelope.
	defaultSource = "admin-service"

	// envelopeVersion is the schema version for EventEnvelope payloads.
	envelopeVersion = 1
)

// EventEnvelope is the standard Kafka message wrapper for all events published by this service.
// It contains routing metadata alongside the typed domain payload.
type EventEnvelope[T any] struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Source        string    `json:"source"`
	Version       int       `json:"version"`
	CorrelationID string    `json:"correlation_id"`
	OccurredAt    time.Time `json:"occurred_at"`
	Data          T         `json:"data"`
}

// EventPublisher defines the adapter contract for publishing domain events to Kafka.
// Implementations must be safe for concurrent use.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, event any) error
}

// EmailKafkaProducer adapts domain EmailSendEvent values into Kafka messages
// using the standard EventEnvelope envelope and dispatches them via a SyncProducer.
type EmailKafkaProducer struct {
	producer sarama.SyncProducer
	source   string
	logger   *slog.Logger
}

// NewEmailKafkaProducer creates an EmailKafkaProducer.
// source is the originating service name embedded in every envelope (e.g. "admin-service").
func NewEmailKafkaProducer(producer sarama.SyncProducer, logger *slog.Logger) *EmailKafkaProducer {
	return &EmailKafkaProducer{
		producer: producer,
		source:   defaultSource,
		logger:   logger,
	}
}

// maskEmail returns a SHA256-truncated 12-char hex string of the email address
// to avoid logging PII while still providing a unique, debuggable fingerprint.
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	h := sha256.New()
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))[:12]
}

// PublishLockoutEmail wraps event in a standard EventEnvelope and publishes it to
// the email.send.v1 Kafka topic. The message key is the recipient email address,
// which guarantees per-user ordering across partitions.
//
// Returns a validation error immediately if event.To is empty.
// Returns a wrapped error if Sarama fails to deliver the message.
// On success, logs the published event at Info level.
func (p *EmailKafkaProducer) PublishLockoutEmail(ctx context.Context, event events.EmailSendEvent) error {
	if event.To == "" {
		return fmt.Errorf("recipient email must not be empty")
	}

	envelope := EventEnvelope[events.EmailSendEvent]{
		ID:            newUUID(),
		Type:          emailSendEventType,
		Source:        p.source,
		Version:       envelopeVersion,
		CorrelationID: correlationIDFromContext(ctx),
		OccurredAt:    time.Now().UTC(),
		Data:          event,
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal lockout email envelope: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: emailSendTopic,
		Key:   sarama.StringEncoder(event.To),
		Value: sarama.ByteEncoder(payload),
	}

	if _, _, err := p.producer.SendMessage(msg); err != nil {
		return fmt.Errorf("publish lockout email: %w", err)
	}

	p.logger.InfoContext(ctx, "lockout email event published to Kafka",
		slog.String("event", "kafka.email_send.published"),
		slog.String("topic", emailSendTopic),
		slog.String("key_masked", maskEmail(event.To)),
	)

	return nil
}

// correlationIDFromContext extracts a correlation ID from context if available,
// returning an empty string when none is set.
// This is a lightweight convention; a proper tracing integration would use
// a context key defined in the platform/tracing package.
func correlationIDFromContext(_ context.Context) string {
	// Future: extract from tracing context (e.g. OpenTelemetry trace ID).
	return ""
}

// newUUID is a thin indirection over generateUUID to keep the adapter
// package free of cross-package imports during unit testing.
// Assigning to a package-level var allows tests to override it if needed.
var newUUID = generateUUID
