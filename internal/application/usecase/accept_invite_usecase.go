package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"

	"github.com/hros/admin-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// NotificationPublisher abstracts the Kafka notification producer.
type NotificationPublisher interface {
	PublishInviteAcceptedNotification(ctx context.Context, event events.NotificationSendEvent) error
}

// AcceptInviteInput contains the validated data required to accept an invite.
type AcceptInviteInput struct {
	Token     string
	Password  string
	IPAddress string
	UserAgent string
}

// AcceptInviteUseCase orchestrates the admin account activation workflow.
type AcceptInviteUseCase struct {
	userRepo        domain.AdminUserRepository
	inviteTokenRepo domain.InviteTokenRepository
	audit           authDomain.AuditLogger
	notificationPub NotificationPublisher
}

// NewAcceptInviteUseCase creates a new AcceptInviteUseCase with all required dependencies.
func NewAcceptInviteUseCase(
	userRepo domain.AdminUserRepository,
	inviteTokenRepo domain.InviteTokenRepository,
	audit authDomain.AuditLogger,
	notificationPub NotificationPublisher,
) *AcceptInviteUseCase {
	return &AcceptInviteUseCase{
		userRepo:        userRepo,
		inviteTokenRepo: inviteTokenRepo,
		audit:           audit,
		notificationPub: notificationPub,
	}
}

// Execute performs the full accept-invite workflow:
//
//  1. Validate the password meets complexity constraints (min 10 chars, 1 upper,
//     1 digit, 1 special character).
//  2. Fetch the InviteToken by the raw token string.
//  3. Guard against expired (> 48 hours) or already-consumed tokens.
//  4. Hash the new password with bcrypt at cost 12.
//  5. Activate the account via AdminUserRepository.ActivateAccount.
//  6. Atomically consume the token via InviteTokenRepository.Consume.
//  7. Emit invite.accepted and admin.activated audit events (best-effort, non-blocking).
//  8. Publish a notification.send Kafka event to the original inviter.
func (uc *AcceptInviteUseCase) Execute(ctx context.Context, input AcceptInviteInput) error {
	if input.Token == "" {
		return errors.New("invite token cannot be empty")
	}
	if input.Password == "" {
		return errors.New("password cannot be empty")
	}

	// Step 1: Validate password complexity.
	if !validatePasswordComplexity(input.Password) {
		return domainErrors.ErrPasswordWeak
	}

	// Step 2: Fetch the invite token.
	inviteToken, err := uc.inviteTokenRepo.FindByToken(ctx, input.Token)
	if err != nil {
		return fmt.Errorf("find invite token: %w", err)
	}

	// Step 3a: Guard — token must not be expired.
	if inviteToken.IsExpired() {
		return domainErrors.ErrInviteExpired
	}

	// Step 3b: Guard — token must not be already used.
	if inviteToken.IsUsed() {
		return domainErrors.ErrInviteUsed
	}

	// Step 4: Hash the new password with bcrypt cost 12.
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	hashedPassword := string(hashedBytes)

	// Step 5: Activate the admin account (update hash + status -> active).
	if err := uc.userRepo.ActivateAccount(ctx, inviteToken.AdminID, hashedPassword); err != nil {
		return fmt.Errorf("activate account: %w", err)
	}

	// Step 6: Atomically consume (mark used) the invite token.
	if _, err := uc.inviteTokenRepo.Consume(ctx, input.Token); err != nil {
		return fmt.Errorf("consume invite token: %w", err)
	}

	now := time.Now().UTC()

	// Step 7a: Emit invite.accepted audit event (best-effort).
	uc.audit.LogInviteAccepted(ctx, events.InviteAcceptedEvent{
		InviteTokenID: inviteToken.ID,
		AdminID:       inviteToken.AdminID,
		InvitedBy:     inviteToken.CreatedBy,
		IPAddress:     input.IPAddress,
		UserAgent:     input.UserAgent,
		OccurredAt:    now,
	})

	// Step 7b: Emit admin.activated audit event (best-effort).
	uc.audit.LogAdminActivated(ctx, events.AdminActivatedEvent{
		AdminID:    inviteToken.AdminID,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		OccurredAt: now,
	})

	// Step 8: Publish in-app notification to the original inviter via Kafka.
	notifEvent := events.NotificationSendEvent{
		RecipientID: inviteToken.CreatedBy,
		Type:        "invite.accepted",
		Title:       "Your invite was accepted",
		Message:     fmt.Sprintf("Admin (ID: %s) has accepted your invitation and activated their account.", inviteToken.AdminID),
		Payload: map[string]interface{}{
			"admin_id":        inviteToken.AdminID,
			"invite_token_id": inviteToken.ID,
		},
		CreatedAt: now,
	}
	if err := uc.notificationPub.PublishInviteAcceptedNotification(ctx, notifEvent); err != nil {
		return fmt.Errorf("publish invite accepted notification: %w", err)
	}

	return nil
}
