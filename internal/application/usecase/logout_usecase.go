package usecase

import (
	"context"
	"errors"
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
func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	if input.Token == "" {
		return fmt.Errorf("session token is empty: %w", domainErrors.ErrTokenNotFound)
	}

	err := uc.sessionRepo.DeleteByToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, domainErrors.ErrTokenNotFound) {
			return fmt.Errorf("delete session token: %w", domainErrors.ErrTokenNotFound)
		}
		return fmt.Errorf("delete session token: %w", err)
	}

	uc.audit.LogLogoutSuccess(ctx, input.Token)
	return nil
}
