package interfaces

import (
	"context"

	"github.com/hros/admin-service/internal/domain/events"
)

// PasswordResetNotifier defines the contract for publishing password reset events (e.g. email notifications).
type PasswordResetNotifier interface {
	// PublishPasswordResetEmail sends a password reset email event.
	PublishPasswordResetEmail(ctx context.Context, event events.EmailSendEvent) error
}
