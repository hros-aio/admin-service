package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hros/admin-service/internal/application/auth"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
)

// CallbackSSOInput represents the input parameters for processing the SSO callback.
type CallbackSSOInput struct {
	Provider  string
	Code      string
	State     string
	IPAddress string
	UserAgent string
}

// CallbackSSOUseCase orchestrates the workflow for handling Identity Provider callback responses.
type CallbackSSOUseCase struct {
	stateCache  interfaces.SSOStateCache
	userRepo    domain.AdminUserRepository
	sessionRepo domain.SessionTokenRepository
	tokens      auth.TokenProvider
	ssoClient   interfaces.SSOClient
	audit       authDomain.AuditLogger
}

// NewCallbackSSOUseCase creates a new CallbackSSOUseCase.
func NewCallbackSSOUseCase(
	stateCache interfaces.SSOStateCache,
	userRepo domain.AdminUserRepository,
	sessionRepo domain.SessionTokenRepository,
	tokens auth.TokenProvider,
	ssoClient interfaces.SSOClient,
	audit authDomain.AuditLogger,
) *CallbackSSOUseCase {
	return &CallbackSSOUseCase{
		stateCache:  stateCache,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokens:      tokens,
		ssoClient:   ssoClient,
		audit:       audit,
	}
}

// Execute handles verifying the CSRF state parameter, exchanging the code,
// checking user mapping constraints, creating local sessions, and emitting audit logs.
func (uc *CallbackSSOUseCase) Execute(ctx context.Context, input CallbackSSOInput) (*LoginOutput, error) {
	if input.Provider == "" {
		return nil, errors.New("provider cannot be empty")
	}
	if input.Code == "" {
		return nil, errors.New("authorization code cannot be empty")
	}
	if input.State == "" {
		return nil, errors.New("state parameter cannot be empty")
	}

	// 1. Verify and atomically consume state parameter
	_, err := uc.stateCache.VerifyAndConsumeState(ctx, input.State)
	if err != nil {
		return nil, domainErrors.ErrInvalidSSOState
	}

	// 2. Exchange code for user profile details
	profile, err := uc.ssoClient.ExchangeCode(ctx, input.Provider, input.Code)
	if err != nil {
		// Log failed SSO login due to exchange error
		auditEvent := events.SSOFailedEvent{
			Provider:   input.Provider,
			Reason:     "code exchange failed",
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			OccurredAt: time.Now().UTC(),
		}
		uc.audit.LogSSOFailed(ctx, auditEvent)
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	if profile == nil {
		// Log failed SSO login due to nil profile
		auditEvent := events.SSOFailedEvent{
			Provider:   input.Provider,
			Reason:     "code exchange returned nil profile",
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			OccurredAt: time.Now().UTC(),
		}
		uc.audit.LogSSOFailed(ctx, auditEvent)
		return nil, errors.New("sso profile is nil")
	}

	// 3. Lookup local admin user account by email or SSO ID
	user, err := uc.userRepo.FindByEmailOrSSO(ctx, profile.Email, profile.Provider, profile.IdentityID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserNotFound) {
			auditEvent := events.SSOFailedEvent{
				Email:      profile.Email,
				Provider:   input.Provider,
				Reason:     "no account linked to this identity",
				IPAddress:  input.IPAddress,
				UserAgent:  input.UserAgent,
				OccurredAt: time.Now().UTC(),
			}
			uc.audit.LogSSOFailed(ctx, auditEvent)
			return nil, domainErrors.ErrNoAccountLinked
		}

		if errors.Is(err, domainErrors.ErrIdentityConflict) {
			auditEvent := events.SSOFailedEvent{
				Email:      profile.Email,
				Provider:   input.Provider,
				Reason:     "identity conflict: email and sso identity map to different users",
				IPAddress:  input.IPAddress,
				UserAgent:  input.UserAgent,
				OccurredAt: time.Now().UTC(),
			}
			uc.audit.LogSSOFailed(ctx, auditEvent)
			return nil, err
		}

		return nil, fmt.Errorf("lookup user: %w", err)
	}

	// 4. Issue RS256 JWT access token (15-min expiry)
	accessToken, err := uc.tokens.GenerateAccessToken(ctx, user, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// 5. Generate session token (refresh token)
	refreshToken, err := uc.tokens.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// 6. Save session token to DB
	session := &domain.SessionToken{
		ID:           domain.NewUUID(),
		AdminID:      user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsPersistent: false,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		CreatedAt:    time.Now(),
	}

	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("save session token: %w", err)
	}

	// 7. Emit login.sso_success audit event
	auditEvent := events.SSOSuccessEvent{
		AdminID:    user.ID,
		Email:      user.Email,
		Provider:   input.Provider,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		OccurredAt: time.Now().UTC(),
	}
	uc.audit.LogSSOSuccess(ctx, auditEvent)

	return &LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: AdminUserSummary{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}
