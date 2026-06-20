package usecase

import (
	"context"
	"fmt"

	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
)

// LogoutUseCase handles the admin logout process.
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
func (uc *LogoutUseCase) Execute(ctx context.Context, token string) error {
	if token == "" {
		return domainErrors.ErrTokenNotFound
	}

	session, err := uc.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("find session token: %w", err)
	}
	if session == nil {
		return domainErrors.ErrTokenNotFound
	}

	if err := uc.sessionRepo.DeleteByToken(ctx, token); err != nil {
		return fmt.Errorf("delete session token: %w", err)
	}

	uc.audit.LogLogoutSuccess(ctx, session.AdminID, token)
	return nil
}
