package interfaces_test

import (
	"context"
	"testing"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain/events"
)

// compilePasswordResetNotifier is a compile-time assertion that the interface is complete
// and can be implemented by a concrete type.
type compilePasswordResetNotifier struct{}

func (c *compilePasswordResetNotifier) PublishPasswordResetEmail(_ context.Context, _ events.EmailSendEvent) error {
	return nil
}

// TestPasswordResetNotifier_InterfaceCompliance verifies the interface is satisfied at compile time.
func TestPasswordResetNotifier_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ interfaces.PasswordResetNotifier = (*compilePasswordResetNotifier)(nil)
}
