package interfaces

import (
	"context"

	"github.com/hros/admin-service/internal/domain/events"
)

// LockoutNotifier defines the contract for publishing lockout notification events.
// Implementations are responsible for dispatching the notification (e.g. via Kafka)
// without coupling the application layer to any specific transport.
type LockoutNotifier interface {
	// PublishLockoutEmail sends a lockout notification email event.
	// Implementations must treat this as best-effort: callers log errors
	// but must not propagate them as login failures.
	PublishLockoutEmail(ctx context.Context, event events.EmailSendEvent) error
}
