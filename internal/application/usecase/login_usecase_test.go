package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mocks ─────────────────────────────────────────────────────────────────────

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Save(ctx context.Context, u *domain.AdminUser) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) Update(ctx context.Context, u *domain.AdminUser) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.AdminUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AdminUser), args.Error(1)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.AdminUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AdminUser), args.Error(1)
}
func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) GetRoleCodeByID(ctx context.Context, roleID string) (string, error) {
	args := m.Called(ctx, roleID)
	return args.String(0), args.Error(1)
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id string, newHash string) error {
	return m.Called(ctx, id, newHash).Error(0)
}

type mockMFACache struct{ mock.Mock }

func (m *mockMFACache) StoreToken(ctx context.Context, token string, adminID string) error {
	return m.Called(ctx, token, adminID).Error(0)
}
func (m *mockMFACache) GetAdminID(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}
func (m *mockMFACache) DeleteToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Save(ctx context.Context, t *domain.SessionToken) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockSessionRepo) FindByToken(ctx context.Context, t string) (*domain.SessionToken, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SessionToken), args.Error(1)
}
func (m *mockSessionRepo) DeleteByToken(ctx context.Context, t string) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockSessionRepo) DeleteByAdminID(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSessionRepo) DeleteAllByAdminID(ctx context.Context, adminID string) error {
	return m.Called(ctx, adminID).Error(0)
}
func (m *mockSessionRepo) Revoke(ctx context.Context, t string, r string) error {
	return m.Called(ctx, t, r).Error(0)
}
func (m *mockSessionRepo) UpdateToken(ctx context.Context, session *domain.SessionToken) error {
	return m.Called(ctx, session).Error(0)
}

type mockPasswordHelper struct{ mock.Mock }

func (m *mockPasswordHelper) Hash(p string) (string, error) { return "", nil }
func (m *mockPasswordHelper) Compare(h, p string) error     { return m.Called(h, p).Error(0) }
func (m *mockPasswordHelper) CompareDummy(p string)         { m.Called(p) }

type mockTokenProvider struct{ mock.Mock }

func (m *mockTokenProvider) GenerateAccessToken(ctx context.Context, u *domain.AdminUser, e time.Duration) (string, error) {
	args := m.Called(ctx, u, e)
	return args.String(0), args.Error(1)
}
func (m *mockTokenProvider) GenerateRefreshToken(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) LogLoginSuccess(ctx context.Context, id, email string) {
	m.Called(ctx, id, email)
}
func (m *mockAuditLogger) LogLoginFailed(ctx context.Context, email, reason string) {
	m.Called(ctx, email, reason)
}
func (m *mockAuditLogger) LogLogoutSuccess(ctx context.Context, token string) { m.Called(ctx, token) }
func (m *mockAuditLogger) LogSessionRefreshed(ctx context.Context, userID string) {
	m.Called(ctx, userID)
}
func (m *mockAuditLogger) LogAccountLocked(ctx context.Context, email string) {
	m.Called(ctx, email)
}
func (m *mockAuditLogger) LogMFAChallengeIssued(ctx context.Context, userID string, email string) {
	m.Called(ctx, userID, email)
}
func (m *mockAuditLogger) LogMFASuccess(ctx context.Context, userID string, email string) {
	m.Called(ctx, userID, email)
}
func (m *mockAuditLogger) LogMFAFailed(ctx context.Context, email string, reason string) {
	m.Called(ctx, email, reason)
}
func (m *mockAuditLogger) LogPasswordResetRequested(ctx context.Context, email string) {
	m.Called(ctx, email)
}

type mockBruteForceCache struct{ mock.Mock }

func (m *mockBruteForceCache) IncrementFailedAttempts(ctx context.Context, email string, window time.Duration) (int, error) {
	args := m.Called(ctx, email, window)
	return args.Int(0), args.Error(1)
}
func (m *mockBruteForceCache) GetFailedAttempts(ctx context.Context, email string) (int, error) {
	args := m.Called(ctx, email)
	return args.Int(0), args.Error(1)
}
func (m *mockBruteForceCache) SetLockout(ctx context.Context, email string, duration time.Duration) error {
	return m.Called(ctx, email, duration).Error(0)
}
func (m *mockBruteForceCache) IsLocked(ctx context.Context, email string) (bool, time.Time, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Get(1).(time.Time), args.Error(2)
}
func (m *mockBruteForceCache) Reset(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}

type mockLockoutNotifier struct{ mock.Mock }

func (m *mockLockoutNotifier) PublishLockoutEmail(ctx context.Context, event events.EmailSendEvent) error {
	return m.Called(ctx, event).Error(0)
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

// newTestUseCase constructs a LoginUseCase with all mocked dependencies.
func newTestUseCase(
	userRepo *mockUserRepo,
	sessionRepo *mockSessionRepo,
	password *mockPasswordHelper,
	tokens *mockTokenProvider,
	audit *mockAuditLogger,
	bruteForce *mockBruteForceCache,
	notifier *mockLockoutNotifier,
) *LoginUseCase {
	// Provide a default fallback expectation for GetRoleCodeByID to return "Admin"
	userRepo.On("GetRoleCodeByID", mock.Anything, mock.Anything).Return("Admin", nil).Maybe()

	return NewLoginUseCase(
		userRepo,
		sessionRepo,
		password,
		tokens,
		audit,
		bruteForce,
		notifier,
		&mockMFACache{},
		slog.Default(),
	)
}

// activeUser returns a standard active admin user for tests.
func activeUser(email string) *domain.AdminUser {
	return &domain.AdminUser{
		ID:           "user-id",
		Email:        email,
		Name:         "Test Admin",
		PasswordHash: "hashed",
		Status:       domain.AdminUserStatusActive,
	}
}

// ─── TSK-AUTH-021: Brute-Force Lockout State Machine Tests ────────────────────

// TestLoginUseCase_AlreadyLocked verifies that when IsLocked returns true,
// ErrAccountLocked is returned immediately without any password or DB work.
func TestLoginUseCase_AlreadyLocked(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "locked@example.com"
	input := LoginInput{Email: email, Password: "any"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	bf.On("IsLocked", ctx, email).Return(true, time.Now().Add(30*time.Minute), nil).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.ErrorIs(t, err, domainErrors.ErrAccountLocked)
	assert.Nil(t, out)

	// None of these must be called.
	userRepo.AssertNotCalled(t, "FindByEmail", mock.Anything, mock.Anything)
	password.AssertNotCalled(t, "Compare", mock.Anything, mock.Anything)
	bf.AssertNotCalled(t, "IncrementFailedAttempts", mock.Anything, mock.Anything, mock.Anything)
	bf.AssertNotCalled(t, "SetLockout", mock.Anything, mock.Anything, mock.Anything)
	bf.AssertNotCalled(t, "Reset", mock.Anything, mock.Anything)
	notifier.AssertNotCalled(t, "PublishLockoutEmail", mock.Anything, mock.Anything)
	audit.AssertNotCalled(t, "LogAccountLocked", mock.Anything, mock.Anything)

	bf.AssertExpectations(t)
}

// TestLoginUseCase_InvalidPassword_LessThan5Failures verifies that an invalid
// password increments the counter but does NOT trigger lockout when count < 5.
func TestLoginUseCase_InvalidPassword_LessThan5Failures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "wrong"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(email)

	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
	password.On("Compare", "hashed", "wrong").Return(errors.New("bcrypt mismatch")).Once()
	audit.On("LogLoginFailed", ctx, email, "invalid password").Once()
	bf.On("IncrementFailedAttempts", ctx, email, failureWindow).Return(3, nil).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
	assert.Nil(t, out)

	// Lockout must NOT be triggered at count = 3.
	bf.AssertNotCalled(t, "SetLockout", mock.Anything, mock.Anything, mock.Anything)
	audit.AssertNotCalled(t, "LogAccountLocked", mock.Anything, mock.Anything)
	notifier.AssertNotCalled(t, "PublishLockoutEmail", mock.Anything, mock.Anything)

	bf.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	password.AssertExpectations(t)
	audit.AssertExpectations(t)
}

// TestLoginUseCase_InvalidPassword_FifthFailure_TriggersLockout verifies that the
// 5th consecutive invalid password triggers SetLockout, an account.locked audit
// event, and a lockout email notification.
func TestLoginUseCase_InvalidPassword_FifthFailure_TriggersLockout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "wrong"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(email)

	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
	password.On("Compare", "hashed", "wrong").Return(errors.New("bcrypt mismatch")).Once()
	audit.On("LogLoginFailed", ctx, email, "invalid password").Once()
	bf.On("IncrementFailedAttempts", ctx, email, failureWindow).Return(5, nil).Once()
	bf.On("SetLockout", ctx, email, lockoutDuration).Return(nil).Once()
	audit.On("LogAccountLocked", ctx, email).Once()
	notifier.On("PublishLockoutEmail", ctx, mock.MatchedBy(func(e events.EmailSendEvent) bool {
		return e.To == email &&
			e.Template == "account_locked_notification" &&
			e.TemplateData["email"] == email
	})).Return(nil).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	// ErrInvalidCredentials is returned; lockout takes effect on the NEXT call (Step 1).
	assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
	assert.Nil(t, out)

	bf.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	password.AssertExpectations(t)
	audit.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

// TestLoginUseCase_ValidPassword_ClearsFailures verifies that a successful login
// resets the brute-force counter and returns a valid session.
func TestLoginUseCase_ValidPassword_ClearsFailures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "correct", IPAddress: "127.0.0.1", UserAgent: "test"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(email)

	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
	password.On("Compare", "hashed", "correct").Return(nil).Once()
	bf.On("Reset", ctx, email).Return(nil).Once()
	tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token", nil).Once()
	tokens.On("GenerateRefreshToken", ctx).Return("refresh-token", nil).Once()
	sessionRepo.On("Save", ctx, mock.MatchedBy(func(s *domain.SessionToken) bool {
		diff := time.Until(s.ExpiresAt)
		return !s.IsPersistent && diff > 23*time.Hour && diff < 25*time.Hour
	})).Return(nil).Once()
	audit.On("LogLoginSuccess", ctx, user.ID, user.Email).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, "access-token", out.AccessToken)
	assert.Equal(t, "refresh-token", out.RefreshToken)

	// Failure-path mocks must not be called.
	bf.AssertNotCalled(t, "IncrementFailedAttempts", mock.Anything, mock.Anything, mock.Anything)
	bf.AssertNotCalled(t, "SetLockout", mock.Anything, mock.Anything, mock.Anything)
	audit.AssertNotCalled(t, "LogAccountLocked", mock.Anything, mock.Anything)
	notifier.AssertNotCalled(t, "PublishLockoutEmail", mock.Anything, mock.Anything)

	bf.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	password.AssertExpectations(t)
	tokens.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

// ─── Pre-existing Tests (retained, updated for new constructor) ───────────────

func TestLoginUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)

	input := LoginInput{
		Email:     "admin@example.com",
		Password:  "password123",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}

	t.Run("Success_NoRememberMe", func(t *testing.T) {
		user := &domain.AdminUser{
			ID:           "user-id",
			Email:        input.Email,
			PasswordHash: "hashed",
			Status:       domain.AdminUserStatusActive,
		}

		inputNoRemember := input
		inputNoRemember.RememberMe = false

		bf.On("IsLocked", ctx, inputNoRemember.Email).Return(false, time.Time{}, nil).Once()
		userRepo.On("FindByEmail", ctx, inputNoRemember.Email).Return(user, nil).Once()
		password.On("Compare", "hashed", inputNoRemember.Password).Return(nil).Once()
		bf.On("Reset", ctx, inputNoRemember.Email).Return(nil).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("refresh-token", nil).Once()
		sessionRepo.On("Save", ctx, mock.MatchedBy(func(s *domain.SessionToken) bool {
			diff := time.Until(s.ExpiresAt)
			return !s.IsPersistent && diff > 23*time.Hour && diff < 25*time.Hour
		})).Return(nil).Once()
		audit.On("LogLoginSuccess", ctx, user.ID, user.Email).Return().Once()

		output, err := uc.Execute(ctx, inputNoRemember)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "access-token", output.AccessToken)
		assert.Equal(t, "refresh-token", output.RefreshToken)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		tokens.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
		bf.AssertExpectations(t)
	})

	t.Run("Success_RememberMe", func(t *testing.T) {
		user := &domain.AdminUser{
			ID:           "user-id",
			Email:        input.Email,
			PasswordHash: "hashed",
			Status:       domain.AdminUserStatusActive,
		}

		inputRemember := input
		inputRemember.RememberMe = true

		bf.On("IsLocked", ctx, inputRemember.Email).Return(false, time.Time{}, nil).Once()
		userRepo.On("FindByEmail", ctx, inputRemember.Email).Return(user, nil).Once()
		password.On("Compare", "hashed", inputRemember.Password).Return(nil).Once()
		bf.On("Reset", ctx, inputRemember.Email).Return(nil).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("refresh-token", nil).Once()
		sessionRepo.On("Save", ctx, mock.MatchedBy(func(s *domain.SessionToken) bool {
			diff := time.Until(s.ExpiresAt)
			return s.IsPersistent && diff > 29*24*time.Hour && diff < 31*24*time.Hour
		})).Return(nil).Once()
		audit.On("LogLoginSuccess", ctx, user.ID, user.Email).Return().Once()

		output, err := uc.Execute(ctx, inputRemember)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "access-token", output.AccessToken)
		assert.Equal(t, "refresh-token", output.RefreshToken)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		tokens.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
		bf.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		bf.On("IsLocked", ctx, input.Email).Return(false, time.Time{}, nil).Once()
		userRepo.On("FindByEmail", ctx, input.Email).Return(nil, domainErrors.ErrUserNotFound).Once()
		password.On("CompareDummy", input.Password).Return().Once()
		audit.On("LogLoginFailed", ctx, input.Email, "user not found").Return().Once()

		output, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		audit.AssertExpectations(t)
		bf.AssertExpectations(t)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		user := &domain.AdminUser{
			ID:           "user-id",
			Email:        input.Email,
			PasswordHash: "hashed",
			Status:       domain.AdminUserStatusActive,
		}

		bf.On("IsLocked", ctx, input.Email).Return(false, time.Time{}, nil).Once()
		userRepo.On("FindByEmail", ctx, input.Email).Return(user, nil).Once()
		password.On("Compare", "hashed", input.Password).Return(errors.New("invalid")).Once()
		audit.On("LogLoginFailed", ctx, input.Email, "invalid password").Return().Once()
		// Count = 2, below threshold — no lockout.
		bf.On("IncrementFailedAttempts", ctx, input.Email, failureWindow).Return(2, nil).Once()

		output, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		audit.AssertExpectations(t)
		bf.AssertExpectations(t)
	})

	t.Run("UserLocked", func(t *testing.T) {
		lockedUntil := time.Now().Add(1 * time.Hour)
		user := &domain.AdminUser{
			ID:          "user-id",
			Email:       input.Email,
			Status:      domain.AdminUserStatusActive,
			LockedUntil: &lockedUntil,
		}

		bf.On("IsLocked", ctx, input.Email).Return(false, time.Time{}, nil).Once()
		userRepo.On("FindByEmail", ctx, input.Email).Return(user, nil).Once()
		password.On("CompareDummy", input.Password).Return().Once()
		audit.On("LogLoginFailed", ctx, input.Email, "user locked").Return().Once()

		output, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, domainErrors.ErrUserLocked)
		assert.Nil(t, output)
		bf.AssertExpectations(t)
	})
}

// ─── Fail-Open Branch Coverage ────────────────────────────────────────────────

// TestLoginUseCase_IsLockedCacheError verifies that when IsLocked returns an error,
// the use case logs a warning and proceeds (fail-open), allowing the login attempt.
func TestLoginUseCase_IsLockedCacheError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "correct", IPAddress: "127.0.0.1", UserAgent: "ua"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(email)

	// Cache error → fail open (locked=false), proceed normally.
	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, errors.New("redis down")).Once()
	userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
	password.On("Compare", "hashed", "correct").Return(nil).Once()
	bf.On("Reset", ctx, email).Return(nil).Once()
	tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("at", nil).Once()
	tokens.On("GenerateRefreshToken", ctx).Return("rt", nil).Once()
	sessionRepo.On("Save", ctx, mock.Anything).Return(nil).Once()
	audit.On("LogLoginSuccess", ctx, user.ID, user.Email).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, out)
	bf.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	password.AssertExpectations(t)
	tokens.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

// TestLoginUseCase_ResetCacheError verifies that when Reset returns an error on a
// successful login, the error is logged (warn) and the login still succeeds (fail-open).
func TestLoginUseCase_ResetCacheError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "correct", IPAddress: "127.0.0.1", UserAgent: "ua"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(email)

	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
	password.On("Compare", "hashed", "correct").Return(nil).Once()
	bf.On("Reset", ctx, email).Return(errors.New("redis down")).Once() // error but non-fatal
	tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("at", nil).Once()
	tokens.On("GenerateRefreshToken", ctx).Return("rt", nil).Once()
	sessionRepo.On("Save", ctx, mock.Anything).Return(nil).Once()
	audit.On("LogLoginSuccess", ctx, user.ID, user.Email).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, out)
	bf.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	password.AssertExpectations(t)
	tokens.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

// TestLoginUseCase_IncrementCacheError verifies that when IncrementFailedAttempts
// returns an error, the use case logs a warning and returns without applying lockout.
func TestLoginUseCase_IncrementCacheError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "wrong"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(email)

	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
	password.On("Compare", "hashed", "wrong").Return(errors.New("bcrypt mismatch")).Once()
	audit.On("LogLoginFailed", ctx, email, "invalid password").Once()
	bf.On("IncrementFailedAttempts", ctx, email, failureWindow).Return(0, errors.New("redis down")).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
	assert.Nil(t, out)

	// Lockout and notification must NOT be triggered when increment fails.
	bf.AssertNotCalled(t, "SetLockout", mock.Anything, mock.Anything, mock.Anything)
	notifier.AssertNotCalled(t, "PublishLockoutEmail", mock.Anything, mock.Anything)
	bf.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	password.AssertExpectations(t)
	audit.AssertExpectations(t)
}

// TestLoginUseCase_SetLockoutCacheError verifies that when SetLockout returns an error,
// the use case logs a warning but still emits the audit event and notification.
func TestLoginUseCase_SetLockoutCacheError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "wrong"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(email)

	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
	password.On("Compare", "hashed", "wrong").Return(errors.New("bcrypt mismatch")).Once()
	audit.On("LogLoginFailed", ctx, email, "invalid password").Once()
	bf.On("IncrementFailedAttempts", ctx, email, failureWindow).Return(5, nil).Once()
	bf.On("SetLockout", ctx, email, lockoutDuration).Return(errors.New("redis down")).Once() // error logged but non-fatal
	audit.On("LogAccountLocked", ctx, email).Once()
	notifier.On("PublishLockoutEmail", ctx, mock.MatchedBy(func(e events.EmailSendEvent) bool {
		return e.To == email
	})).Return(nil).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
	assert.Nil(t, out)
	bf.AssertExpectations(t)
	audit.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

// TestLoginUseCase_PublishLockoutEmailError verifies that when PublishLockoutEmail
// returns an error, it is logged but NOT propagated — ErrInvalidCredentials is returned.
func TestLoginUseCase_PublishLockoutEmailError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "wrong"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(email)

	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
	password.On("Compare", "hashed", "wrong").Return(errors.New("bcrypt mismatch")).Once()
	audit.On("LogLoginFailed", ctx, email, "invalid password").Once()
	bf.On("IncrementFailedAttempts", ctx, email, failureWindow).Return(5, nil).Once()
	bf.On("SetLockout", ctx, email, lockoutDuration).Return(nil).Once()
	audit.On("LogAccountLocked", ctx, email).Once()
	notifier.On("PublishLockoutEmail", ctx, mock.MatchedBy(func(e events.EmailSendEvent) bool {
		return e.To == email
	})).Return(errors.New("kafka unavailable")).Once() // error must NOT propagate

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	// Kafka error must not change the outcome.
	assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
	assert.Nil(t, out)
	bf.AssertExpectations(t)
	audit.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

// TestLoginUseCase_FindByEmailDBError verifies that a non-user-not-found repository
// error is wrapped and returned as an internal error.
func TestLoginUseCase_FindByEmailDBError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "admin@example.com"
	input := LoginInput{Email: email, Password: "any"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	dbErr := errors.New("connection timeout")
	bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, email).Return(nil, dbErr).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "find user by email")
	assert.Nil(t, out)
	userRepo.AssertExpectations(t)
	bf.AssertExpectations(t)
}

// TestLoginUseCase_EmailNormalization verifies that input emails are normalized
// (lowercased and whitespace trimmed) before any database, cache, or audit operations.
func TestLoginUseCase_EmailNormalization(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rawEmail := "  ADMIN@Example.Com  "
	normalizedEmail := "admin@example.com"
	input := LoginInput{Email: rawEmail, Password: "correct", IPAddress: "127.0.0.1", UserAgent: "ua"}

	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)
	bf := new(mockBruteForceCache)
	notifier := new(mockLockoutNotifier)

	user := activeUser(normalizedEmail)

	bf.On("IsLocked", ctx, normalizedEmail).Return(false, time.Time{}, nil).Once()
	userRepo.On("FindByEmail", ctx, normalizedEmail).Return(user, nil).Once()
	password.On("Compare", "hashed", "correct").Return(nil).Once()
	bf.On("Reset", ctx, normalizedEmail).Return(nil).Once()
	tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("at", nil).Once()
	tokens.On("GenerateRefreshToken", ctx).Return("rt", nil).Once()
	sessionRepo.On("Save", ctx, mock.Anything).Return(nil).Once()
	audit.On("LogLoginSuccess", ctx, user.ID, normalizedEmail).Once()

	uc := newTestUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier)
	out, err := uc.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, out)
	bf.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	password.AssertExpectations(t)
	tokens.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

// TestLoginUseCase_SuperAdminMFA verifies that logging in as a Super Admin
// intercepts standard JWT token generation and issues an intermediate MFA challenge token instead.
func TestLoginUseCase_SuperAdminMFA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "superadmin@example.com"
	input := LoginInput{Email: email, Password: "correct", IPAddress: "127.0.0.1", UserAgent: "ua"}

	t.Run("successfully intercepts Super Admin login and returns MFA token", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		password := new(mockPasswordHelper)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		bf := new(mockBruteForceCache)
		notifier := new(mockLockoutNotifier)
		mfaCache := new(mockMFACache)

		user := activeUser(email)
		user.RoleID = "super-admin-role-id"

		bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
		userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
		password.On("Compare", "hashed", "correct").Return(nil).Once()
		bf.On("Reset", ctx, email).Return(nil).Once()
		userRepo.On("GetRoleCodeByID", ctx, user.RoleID).Return("SUPER_ADMIN", nil).Once()
		mfaCache.On("StoreToken", ctx, mock.Anything, user.ID).Return(nil).Once()
		audit.On("LogMFAChallengeIssued", ctx, user.ID, user.Email).Once()

		uc := NewLoginUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, out.MFARequired)
		assert.Regexp(t, "^[0-9a-fA-F]{64}$", out.MFAToken)
		assert.Equal(t, []string{"totp", "webauthn"}, out.MFAMethods)
		assert.Empty(t, out.AccessToken)
		assert.Empty(t, out.RefreshToken)

		// Assert that token provider and session saves were NOT called
		tokens.AssertNotCalled(t, "GenerateAccessToken", mock.Anything, mock.Anything, mock.Anything)
		tokens.AssertNotCalled(t, "GenerateRefreshToken", mock.Anything)
		sessionRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)

		bf.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		mfaCache.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("fails Super Admin login if MFACache fails to store token", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		password := new(mockPasswordHelper)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		bf := new(mockBruteForceCache)
		notifier := new(mockLockoutNotifier)
		mfaCache := new(mockMFACache)

		user := activeUser(email)
		user.RoleID = "super-admin-role-id"

		bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
		userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
		password.On("Compare", "hashed", "correct").Return(nil).Once()
		bf.On("Reset", ctx, email).Return(nil).Once()
		userRepo.On("GetRoleCodeByID", ctx, user.RoleID).Return("SUPER_ADMIN", nil).Once()

		cacheErr := errors.New("redis connection timeout")
		mfaCache.On("StoreToken", ctx, mock.Anything, user.ID).Return(cacheErr).Once()

		uc := NewLoginUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "store mfa token")
		assert.Nil(t, out)

		bf.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		mfaCache.AssertExpectations(t)
	})

	t.Run("fails Super Admin login if role check fails", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		password := new(mockPasswordHelper)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		bf := new(mockBruteForceCache)
		notifier := new(mockLockoutNotifier)
		mfaCache := new(mockMFACache)

		user := activeUser(email)
		user.RoleID = "super-admin-role-id"

		bf.On("IsLocked", ctx, email).Return(false, time.Time{}, nil).Once()
		userRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
		password.On("Compare", "hashed", "correct").Return(nil).Once()
		bf.On("Reset", ctx, email).Return(nil).Once()

		dbErr := errors.New("db error")
		userRepo.On("GetRoleCodeByID", ctx, user.RoleID).Return("", dbErr).Once()

		uc := NewLoginUseCase(userRepo, sessionRepo, password, tokens, audit, bf, notifier, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "get role code")
		assert.Nil(t, out)

		bf.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
	})
}
