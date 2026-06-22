package producer

import "github.com/google/uuid"

// generateUUID returns a new random UUID string.
// Isolated here so email_events.go can be tested without importing internal/domain.
func generateUUID() string {
	return uuid.New().String()
}
