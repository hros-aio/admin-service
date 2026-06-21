// Package usecase defines application layer orchestration workflows and orchestrators.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hros/admin-service/internal/application/auth"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
)

// RefreshInput represents the input for the refresh session use case.
type RefreshInput struct {
	RefreshToken string
}

// RefreshOutput represents the output of the refresh session use case.
type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

// RefreshSessionUseCase orchestrates the refresh token rotation process.
type RefreshSessionUseCase struct {
	userRepo    domain.AdminUserRepository
	sessionRepo domain.SessionTokenRepository
	tokens      auth.TokenProvider
	audit       authDomain.AuditLogger
}

// NewRefreshSessionUseCase creates a new RefreshSessionUseCase.
func NewRefreshSessionUseCase(
	userRepo domain.AdminUserRepository,
	sessionRepo domain.SessionTokenRepository,
	tokens auth.TokenProvider,
	audit authDomain.AuditLogger,
) *RefreshSessionUseCase {
	return &RefreshSessionUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokens:      tokens,
		audit:       audit,
	}
}

// Execute handles the refresh token rotation business logic.
func (uc *RefreshSessionUseCase) Execute(ctx context.Context, input RefreshInput) (*RefreshOutput, error) {
	if input.RefreshToken == "" {
		return nil, domainErrors.ErrInvalidRefreshToken
	}

	// 1. Fetch the session token by its value
	session, err := uc.sessionRepo.FindByToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, domainErrors.ErrInvalidRefreshToken
	}
	if session == nil {
		return nil, domainErrors.ErrInvalidRefreshToken
	}

	// 2. Check if the token is revoked
	if session.IsRevoked() {
		return nil, domainErrors.ErrInvalidRefreshToken
	}

	// 3. Check if the token is expired
	if session.IsExpired() {
		return nil, domainErrors.ErrTokenExpired
	}

	// 4. Fetch the admin user to check status and generate access token claims
	user, err := uc.userRepo.FindByID(ctx, session.AdminID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserNotFound) {
			return nil, domainErrors.ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	// 5. Verify user status
	if !user.IsActive() {
		return nil, domainErrors.ErrUserInactive
	}
	if user.IsLocked() {
		return nil, domainErrors.ErrUserLocked
	}

	// 6. Generate a new RS256 JWT access token (15-min expiry)
	newAccessToken, err := uc.tokens.GenerateAccessToken(ctx, user, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// 7. Generate a new refresh token string
	newRefreshToken, err := uc.tokens.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// 8. Rotate session model state in-place and save to database
	expiryDuration := 24 * time.Hour
	if session.IsPersistent {
		expiryDuration = 30 * 24 * time.Hour
	}
	newExpiry := time.Now().Add(expiryDuration)
	if _, err := session.Rotate(newExpiry); err != nil {
		return nil, fmt.Errorf("rotate session token: %w", err)
	}
	session.RefreshToken = newRefreshToken

	if err := uc.sessionRepo.UpdateToken(ctx, session); err != nil {
		return nil, fmt.Errorf("update session token: %w", err)
	}

	// 9. Log the refresh session audit event
	uc.audit.LogSessionRefreshed(ctx, user.ID)

	return &RefreshOutput{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
