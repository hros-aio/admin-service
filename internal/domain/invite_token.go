package domain

import (
	"context"
	"time"
)

// InviteToken represents a secure invitation token issued to a new administrator.
type InviteToken struct {
	ID        string
	AdminID   string
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedBy string
	CreatedAt time.Time
}

// IsExpired checks if the invite token has expired (now >= ExpiresAt).
func (it *InviteToken) IsExpired() bool {
	return !time.Now().Before(it.ExpiresAt)
}

// IsUsed checks if the invite token has already been consumed.
func (it *InviteToken) IsUsed() bool {
	return it.UsedAt != nil
}

// Consume marks the invite token as used by setting UsedAt to the current time,
// but only if it has not already been consumed (preserving the original redemption time).
func (it *InviteToken) Consume() {
	if it.UsedAt == nil {
		now := time.Now()
		it.UsedAt = &now
	}
}

// InviteTokenRepository defines the interface for persisting and retrieving invite tokens.
type InviteTokenRepository interface {
	Save(ctx context.Context, token *InviteToken) error
	FindByToken(ctx context.Context, token string) (*InviteToken, error)
	Update(ctx context.Context, token *InviteToken) error
	Consume(ctx context.Context, token string) (*InviteToken, error)
}
