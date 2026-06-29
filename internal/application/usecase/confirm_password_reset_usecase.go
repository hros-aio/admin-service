package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
	"golang.org/x/crypto/bcrypt"
)

// ConfirmPasswordResetInput represents the input for confirming a password reset.
type ConfirmPasswordResetInput struct {
	Token     string
	Password  string
	IPAddress string
	UserAgent string
}

// ConfirmPasswordResetUseCase orchestrates the workflow for confirming a password reset.
type ConfirmPasswordResetUseCase struct {
	userRepo    domain.AdminUserRepository
	sessionRepo domain.SessionTokenRepository
	resetCache  interfaces.PasswordResetCache
	audit       authDomain.AuditLogger
}

// NewConfirmPasswordResetUseCase creates a new ConfirmPasswordResetUseCase.
func NewConfirmPasswordResetUseCase(
	userRepo domain.AdminUserRepository,
	sessionRepo domain.SessionTokenRepository,
	resetCache interfaces.PasswordResetCache,
	audit authDomain.AuditLogger,
) *ConfirmPasswordResetUseCase {
	return &ConfirmPasswordResetUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		resetCache:  resetCache,
		audit:       audit,
	}
}

// Execute performs the confirm password reset workflow.
func (uc *ConfirmPasswordResetUseCase) Execute(ctx context.Context, input ConfirmPasswordResetInput) error {
	if input.Token == "" {
		return errors.New("reset token cannot be empty")
	}
	if input.Password == "" {
		return errors.New("password cannot be empty")
	}

	// 1. Validate password meets complexity constraints (min 10 chars, 1 upper, 1 number, 1 special).
	// Use utf8.RuneCountInString to count runes (characters) rather than bytes.
	if !validatePasswordComplexity(input.Password) {
		return domainErrors.ErrPasswordWeak
	}

	// 2. Atomically consume the token from cache before any account mutation.
	adminID, err := uc.resetCache.ConsumeToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, domainErrors.ErrTokenExpired) {
			return domainErrors.ErrTokenExpired
		}
		if errors.Is(err, domainErrors.ErrTokenUsed) {
			return domainErrors.ErrTokenUsed
		}
		return fmt.Errorf("consume reset token: %w", err)
	}

	// 3. Find the admin user to get the email address for audit logging.
	user, err := uc.userRepo.FindByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	// 4. Hash new password with bcrypt (cost 12).
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	hashedPassword := string(hashedBytes)

	// 5. Save the hashed password via repository.
	if err := uc.userRepo.UpdatePassword(ctx, adminID, hashedPassword); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// 6. Delete all active sessions to force re-authentication everywhere.
	if err := uc.sessionRepo.DeleteAllByAdminID(ctx, adminID); err != nil {
		return fmt.Errorf("delete session tokens: %w", err)
	}

	// 7. Emit password.reset_completed to the audit log.
	completedEvent := events.PasswordResetCompletedEvent{
		Email:      user.Email,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		OccurredAt: time.Now().UTC(),
	}
	uc.audit.LogPasswordResetCompleted(ctx, completedEvent)

	return nil
}

// validatePasswordComplexity checks if the password meets:
// - minimum 10 characters (runes)
// - at least 1 uppercase letter
// - at least 1 digit
// - at least 1 special character
func validatePasswordComplexity(password string) bool {
	if utf8.RuneCountInString(password) < 10 {
		return false
	}
	var hasUpper, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsSpace(r):
			hasSpecial = true
		}
	}
	return hasUpper && hasDigit && hasSpecial
}
