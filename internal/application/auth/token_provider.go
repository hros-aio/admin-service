package auth

import (
	"context"
	"time"

	"github.com/hros/admin-service/internal/domain"
)

// TokenProvider defines the interface for generating and validating tokens.
type TokenProvider interface {
	// GenerateAccessToken generates an RS256 JWT access token for an admin user.
	GenerateAccessToken(ctx context.Context, user *domain.AdminUser, expiry time.Duration) (string, error)
	// GenerateRefreshToken generates a secure random refresh token.
	GenerateRefreshToken(ctx context.Context) (string, error)
}
