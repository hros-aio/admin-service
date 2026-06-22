package interfaces_test

import (
	"context"
	"testing"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain/events"
)

// compileLockoutNotifier is a compile-time assertion that the interface is complete
// and can be implemented by a concrete type.
type compileLockoutNotifier struct{}

func (c *compileLockoutNotifier) PublishLockoutEmail(_ context.Context, _ events.EmailSendEvent) error {
	return nil
}

// TestLockoutNotifier_InterfaceCompliance verifies the interface is satisfied at compile time.
func TestLockoutNotifier_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ interfaces.LockoutNotifier = (*compileLockoutNotifier)(nil)
}
