package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hros/admin-service/internal/application/auth"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/pquerna/otp/totp"
)

// VerifyMFAInput represents the input for the MFA verification use case.
type VerifyMFAInput struct {
	MFAToken   string
	Method     string
	Code       string
	RememberMe bool
	IPAddress  string
	UserAgent  string
}

// VerifyMFAOutput represents the output returned on successful MFA verification.
type VerifyMFAOutput struct {
	AccessToken  string
	RefreshToken string
	User         AdminUserSummary
}

// VerifyMFAUseCase handles second-factor MFA verification for admin logins.
type VerifyMFAUseCase struct {
	userRepo    domain.AdminUserRepository
	sessionRepo domain.SessionTokenRepository
	tokens      auth.TokenProvider
	audit       authDomain.AuditLogger
	mfaCache    interfaces.MFACache
	logger      *slog.Logger
}

// NewVerifyMFAUseCase creates a new VerifyMFAUseCase.
func NewVerifyMFAUseCase(
	userRepo domain.AdminUserRepository,
	sessionRepo domain.SessionTokenRepository,
	tokens auth.TokenProvider,
	audit authDomain.AuditLogger,
	mfaCache interfaces.MFACache,
	logger *slog.Logger,
) *VerifyMFAUseCase {
	return &VerifyMFAUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokens:      tokens,
		audit:       audit,
		mfaCache:    mfaCache,
		logger:      logger,
	}
}

// Execute validates the second-factor code (TOTP) and issues JWT tokens on success.
func (uc *VerifyMFAUseCase) Execute(ctx context.Context, input VerifyMFAInput) (*VerifyMFAOutput, error) {
	// 1. Fetch admin ID from MFACache
	adminID, err := uc.mfaCache.GetAdminID(ctx, input.MFAToken)
	if err != nil {
		if errors.Is(err, domainErrors.ErrMFATokenExpired) {
			return nil, domainErrors.ErrMFATokenExpired
		}
		return nil, fmt.Errorf("failed to retrieve admin ID from cache: %w", err)
	}
	if adminID == "" {
		return nil, domainErrors.ErrMFATokenExpired
	}

	// 2. Fetch admin user
	user, err := uc.userRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	// Double check user status
	if !user.IsActive() {
		return nil, domainErrors.ErrUserInactive
	}
	if user.IsLocked() {
		return nil, domainErrors.ErrUserLocked
	}

	// 3. Validate code based on method
	switch input.Method {
	case "totp":
		if user.TotpSecret == "" {
			uc.audit.LogMFAFailed(ctx, user.Email, "TOTP is not configured")
			return nil, domainErrors.ErrMFAInvalid
		}
		if !totp.Validate(input.Code, user.TotpSecret) {
			uc.audit.LogMFAFailed(ctx, user.Email, "invalid TOTP code")
			return nil, domainErrors.ErrMFAInvalid
		}
	default:
		uc.audit.LogMFAFailed(ctx, user.Email, fmt.Sprintf("unsupported method: %s", input.Method))
		return nil, fmt.Errorf("unsupported verification method: %s", input.Method)
	}

	// 4. Verification success! Log audit event
	uc.audit.LogMFASuccess(ctx, user.ID, user.Email)

	// 5. Generate JWT tokens
	accessToken, err := uc.tokens.GenerateAccessToken(ctx, user, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := uc.tokens.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// 6. Save persistent session token to repository
	expiryDuration := 24 * time.Hour
	if input.RememberMe {
		expiryDuration = 30 * 24 * time.Hour
	}

	session := &domain.SessionToken{
		ID:           domain.NewUUID(),
		AdminID:      user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(expiryDuration),
		IsPersistent: input.RememberMe,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		CreatedAt:    time.Now(),
	}

	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("save session token: %w", err)
	}

	// 7. Delete MFA token from cache
	if err := uc.mfaCache.DeleteToken(ctx, input.MFAToken); err != nil {
		// Log error but do not fail the request since authentication succeeded
		uc.logger.WarnContext(ctx, "failed to delete MFA token from cache after success",
			slog.String("event", "verify_mfa_usecase.delete_mfa_token_failed"),
			slog.Any("error", err),
		)
	}

	return &VerifyMFAOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: AdminUserSummary{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}
