package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
)

// LogoutInput represents the input for the logout use case.
type LogoutInput struct {
	RefreshToken string
}

// LogoutUseCase handles the admin logout process by revoking/deleting the session token.
type LogoutUseCase struct {
	sessionRepo domain.SessionTokenRepository
	audit       authDomain.AuditLogger
}

// NewLogoutUseCase creates a new LogoutUseCase.
func NewLogoutUseCase(
	sessionRepo domain.SessionTokenRepository,
	audit authDomain.AuditLogger,
) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
		audit:       audit,
	}
}

// Execute performs the logout flow.
func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	if input.RefreshToken == "" {
		return errors.New("refresh token cannot be empty")
	}

	// 1. Delete the token from session_tokens via repository.
	// Since GORM Delete is idempotent and returns nil if the record is not found,
	// this works for both existing and non-existing tokens.
	if err := uc.sessionRepo.DeleteByToken(ctx, input.RefreshToken); err != nil {
		return fmt.Errorf("failed to delete session token: %w", err)
	}

	// 2. Emit logout.success to audit log interface
	uc.audit.LogLogoutSuccess(ctx, input.RefreshToken)

	return nil
}
