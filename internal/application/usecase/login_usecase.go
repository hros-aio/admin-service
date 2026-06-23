package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hros/admin-service/internal/application/auth"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
)

const (
	// maxFailedAttempts is the number of consecutive failures that triggers a lockout.
	maxFailedAttempts = 5

	// failureWindow is the sliding window for counting consecutive failed attempts.
	failureWindow = 15 * time.Minute

	// lockoutDuration is the duration an account is locked after reaching maxFailedAttempts.
	lockoutDuration = 30 * time.Minute
)

// LoginUseCase handles the admin login process, including brute-force lockout defense.
type LoginUseCase struct {
	userRepo      domain.AdminUserRepository
	sessionRepo   domain.SessionTokenRepository
	password      auth.PasswordHelper
	tokens        auth.TokenProvider
	audit         authDomain.AuditLogger
	bruteForce    interfaces.BruteForceCache
	lockoutNotify interfaces.LockoutNotifier
	mfaCache      interfaces.MFACache
	logger        *slog.Logger
}

// NewLoginUseCase creates a new LoginUseCase with all required dependencies.
func NewLoginUseCase(
	userRepo domain.AdminUserRepository,
	sessionRepo domain.SessionTokenRepository,
	password auth.PasswordHelper,
	tokens auth.TokenProvider,
	audit authDomain.AuditLogger,
	bruteForce interfaces.BruteForceCache,
	lockoutNotify interfaces.LockoutNotifier,
	mfaCache interfaces.MFACache,
	logger *slog.Logger,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		password:      password,
		tokens:        tokens,
		audit:         audit,
		bruteForce:    bruteForce,
		lockoutNotify: lockoutNotify,
		mfaCache:      mfaCache,
		logger:        logger,
	}
}

// Execute performs the login flow with brute-force lockout defense.
//
// Step 1: Check the brute-force cache. If the account is already locked,
// return ErrAccountLocked immediately — no password verification performed.
//
// Step 2: On invalid password, increment the failure counter. If the counter
// reaches maxFailedAttempts, apply SetLockout, emit an account.locked audit
// event, and publish a best-effort email notification. Still return
// ErrInvalidCredentials (the lockout takes effect on the next call via Step 1).
//
// Step 3: On successful password verification, reset the failure counter.
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// ── Step 1: Brute-force pre-check ─────────────────────────────────────────
	// Check lockout state BEFORE any expensive operation (bcrypt, DB lookup).
	// Fail-open: if the cache is unavailable, allow the request to proceed.
	locked, _, cacheErr := uc.bruteForce.IsLocked(ctx, email)
	if cacheErr != nil {
		uc.logger.WarnContext(ctx, "brute force cache unavailable during IsLocked; failing open",
			slog.String("error", cacheErr.Error()),
		)
	}
	if locked {
		return nil, domainErrors.ErrAccountLocked
	}

	// 1b. Fetch user by email
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserNotFound) {
			// Perform dummy comparison to prevent timing attacks.
			uc.password.CompareDummy(input.Password)
			uc.audit.LogLoginFailed(ctx, email, "user not found")
			return nil, domainErrors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	// 1c. Check if user is active
	if !user.IsActive() {
		uc.password.CompareDummy(input.Password) // Keep timing consistent
		uc.audit.LogLoginFailed(ctx, email, "user inactive")
		return nil, domainErrors.ErrUserInactive
	}

	// 1d. Check if user is locked at the domain level (LockedUntil field)
	if user.IsLocked() {
		uc.password.CompareDummy(input.Password) // Keep timing consistent
		uc.audit.LogLoginFailed(ctx, email, "user locked")
		return nil, domainErrors.ErrUserLocked
	}

	// ── Step 2: Password verification ─────────────────────────────────────────
	if err := uc.password.Compare(user.PasswordHash, input.Password); err != nil {
		uc.audit.LogLoginFailed(ctx, email, "invalid password")
		uc.handleFailedAttempt(ctx, email, user)
		return nil, domainErrors.ErrInvalidCredentials
	}

	// ── Step 3: Successful login — reset the failure counter ──────────────────
	if resetErr := uc.bruteForce.Reset(ctx, email); resetErr != nil {
		uc.logger.WarnContext(ctx, "brute force cache unavailable during Reset; proceeding",
			slog.String("error", resetErr.Error()),
		)
	}

	roleName, err := uc.userRepo.GetRoleNameByID(ctx, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("get role name: %w", err)
	}

	if roleName == "Super Admin" {
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return nil, fmt.Errorf("generate mfa token: %w", err)
		}
		mfaToken := hex.EncodeToString(tokenBytes)

		if err := uc.mfaCache.StoreToken(ctx, mfaToken, user.ID); err != nil {
			uc.logger.ErrorContext(ctx, "failed to cache MFA token for Super Admin",
				slog.String("event", "login_usecase.cache_mfa_token_failed"),
				slog.String("user_id", user.ID),
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("store mfa token: %w", err)
		}

		uc.logger.InfoContext(ctx, "Super Admin login intercepted; MFA token generated",
			slog.String("event", "login_usecase.mfa_token_generated"),
			slog.String("user_id", user.ID),
			slog.String("key", "auth:mfa_token:[REDACTED]"),
		)

		return &LoginOutput{
			MFARequired: true,
			MFAToken:    mfaToken,
			MFAMethods:  []string{"totp", "webauthn"},
			User: AdminUserSummary{
				ID:    user.ID,
				Email: user.Email,
				Name:  user.Name,
			},
		}, nil
	}

	// Issue RS256 JWT access token (15-min expiry)
	accessToken, err := uc.tokens.GenerateAccessToken(ctx, user, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// Generate session token (refresh token)
	refreshToken, err := uc.tokens.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Save session token to DB
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

	// Emit login.success
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

// handleFailedAttempt increments the failure counter and, if the threshold is reached,
// applies a cache lockout, emits the account.locked audit event, and publishes a
// best-effort email notification. All cache errors are logged and treated as fail-open.
func (uc *LoginUseCase) handleFailedAttempt(ctx context.Context, email string, user *domain.AdminUser) {
	count, incrErr := uc.bruteForce.IncrementFailedAttempts(ctx, email, failureWindow)
	if incrErr != nil {
		uc.logger.WarnContext(ctx, "brute force cache unavailable during IncrementFailedAttempts; failing open",
			slog.String("error", incrErr.Error()),
		)
		return
	}

	if count < maxFailedAttempts {
		return
	}

	// Threshold reached — apply lockout.
	if lockErr := uc.bruteForce.SetLockout(ctx, email, lockoutDuration); lockErr != nil {
		uc.logger.WarnContext(ctx, "brute force cache unavailable during SetLockout",
			slog.String("error", lockErr.Error()),
		)
	}

	// Emit account.locked audit event.
	uc.audit.LogAccountLocked(ctx, email)

	// Publish best-effort email notification (errors logged, never propagated).
	unlockAt := time.Now().UTC().Add(lockoutDuration)
	emailEvent := events.EmailSendEvent{
		To:       user.Email,
		Subject:  "Your account has been temporarily locked",
		Template: "account_locked_notification",
		TemplateData: map[string]interface{}{
			"email":     user.Email,
			"unlock_at": unlockAt.Format(time.RFC3339),
		},
	}

	if pubErr := uc.lockoutNotify.PublishLockoutEmail(ctx, emailEvent); pubErr != nil {
		uc.logger.ErrorContext(ctx, "failed to publish lockout email notification",
			slog.String("error", pubErr.Error()),
		)
	}
}
