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

// LoginUseCase handles the admin login process.
type LoginUseCase struct {
	userRepo     domain.AdminUserRepository
	sessionRepo  domain.SessionTokenRepository
	password     auth.PasswordHelper
	tokens       auth.TokenProvider
	audit        authDomain.AuditLogger
}

// NewLoginUseCase creates a new LoginUseCase.
func NewLoginUseCase(
	userRepo domain.AdminUserRepository,
	sessionRepo domain.SessionTokenRepository,
	password auth.PasswordHelper,
	tokens auth.TokenProvider,
	audit authDomain.AuditLogger,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		password:    password,
		tokens:      tokens,
		audit:       audit,
	}
}

// Execute performs the login flow.
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// 1. Fetch user by email
	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserNotFound) {
			// 2a. User not found: Perform dummy comparison to prevent timing attacks
			uc.password.CompareDummy(input.Password)
			uc.audit.LogLoginFailed(ctx, input.Email, "user not found")
			return nil, domainErrors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	// 2b. Check if user is active
	if !user.IsActive() {
		uc.password.CompareDummy(input.Password) // Keep timing consistent
		uc.audit.LogLoginFailed(ctx, input.Email, "user inactive")
		return nil, domainErrors.ErrUserInactive
	}

	// 2c. Check if user is locked
	if user.IsLocked() {
		uc.password.CompareDummy(input.Password) // Keep timing consistent
		uc.audit.LogLoginFailed(ctx, input.Email, "user locked")
		return nil, domainErrors.ErrUserLocked
	}

	// 3. Verify password
	if err := uc.password.Compare(user.PasswordHash, input.Password); err != nil {
		uc.audit.LogLoginFailed(ctx, input.Email, "invalid password")
		return nil, domainErrors.ErrInvalidCredentials
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
		ID:           domain.NewUUID(), // Assuming a helper exists or we'll generate it
		AdminID:      user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour), // 30 days expiry for refresh token
		IsPersistent: true,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		CreatedAt:    time.Now(),
	}

	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("save session token: %w", err)
	}

	// 7. Emit login.success
	uc.audit.LogLoginSuccess(ctx, user.ID, user.Email)

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
