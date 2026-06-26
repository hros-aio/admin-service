package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
)

// RequestPasswordResetInput represents the input for requesting a password reset.
type RequestPasswordResetInput struct {
	Email     string
	IPAddress string
	UserAgent string
}

// RequestPasswordResetUseCase orchestrates the workflow for requesting a self-service password reset.
type RequestPasswordResetUseCase struct {
	userRepo   domain.AdminUserRepository
	resetCache interfaces.PasswordResetCache
	audit      authDomain.AuditLogger
	notifier   interfaces.PasswordResetNotifier
}

// NewRequestPasswordResetUseCase creates a new RequestPasswordResetUseCase.
func NewRequestPasswordResetUseCase(
	userRepo domain.AdminUserRepository,
	resetCache interfaces.PasswordResetCache,
	audit authDomain.AuditLogger,
	notifier interfaces.PasswordResetNotifier,
) *RequestPasswordResetUseCase {
	return &RequestPasswordResetUseCase{
		userRepo:   userRepo,
		resetCache: resetCache,
		audit:      audit,
		notifier:   notifier,
	}
}

// Execute handles the request password reset flow.
func (uc *RequestPasswordResetUseCase) Execute(ctx context.Context, input RequestPasswordResetInput) error {
	if input.Email == "" {
		return errors.New("email cannot be empty")
	}

	// 1. Find user via repository.
	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserNotFound) {
			// Always return success immediately if the email is not found, to prevent email enumeration.
			return nil
		}
		return fmt.Errorf("find user: %w", err)
	}

	// 2. Generate a secure single-use token (32 bytes = 256 bits of entropy).
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generate secure token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// 3. Store token via PasswordResetCache with a strict 60-minute TTL.
	ttl := 60 * time.Minute
	if err := uc.resetCache.StoreToken(ctx, token, user.ID, ttl); err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}

	// 4. Emit password.reset_requested to the audit log.
	auditEvent := events.PasswordResetRequestedEvent{
		Email:      user.Email,
		Token:      token,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		OccurredAt: time.Now().UTC(),
	}
	uc.audit.LogPasswordResetRequested(ctx, auditEvent)

	// 5. Publish the email.send Kafka event.
	emailEvent := events.EmailSendEvent{
		To:       user.Email,
		Subject:  "Reset your password",
		Template: "password_reset_request",
		TemplateData: map[string]interface{}{
			"email": user.Email,
			"token": token,
		},
	}
	if err := uc.notifier.PublishPasswordResetEmail(ctx, emailEvent); err != nil {
		// Rollback stored token from cache on failure to publish event.
		_ = uc.resetCache.DeleteToken(ctx, token)
		return fmt.Errorf("publish password reset email: %w", err)
	}

	return nil
}
