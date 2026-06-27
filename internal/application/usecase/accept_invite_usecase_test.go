package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
)

// ---------------------------------------------------------------------------
// Mock: InviteTokenRepository
// ---------------------------------------------------------------------------

type mockInviteTokenRepository struct{ mock.Mock }

func (m *mockInviteTokenRepository) Save(ctx context.Context, token *domain.InviteToken) error {
	return m.Called(ctx, token).Error(0)
}
func (m *mockInviteTokenRepository) FindByToken(ctx context.Context, token string) (*domain.InviteToken, error) {
	args := m.Called(ctx, token)
	if v := args.Get(0); v != nil {
		return v.(*domain.InviteToken), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockInviteTokenRepository) Update(ctx context.Context, token *domain.InviteToken) error {
	return m.Called(ctx, token).Error(0)
}
func (m *mockInviteTokenRepository) Consume(ctx context.Context, token string) (*domain.InviteToken, error) {
	args := m.Called(ctx, token)
	if v := args.Get(0); v != nil {
		return v.(*domain.InviteToken), args.Error(1)
	}
	return nil, args.Error(1)
}

// ---------------------------------------------------------------------------
// Mock: AdminUserRepository (only ActivateAccount is exercised here)
// ---------------------------------------------------------------------------

type mockAdminUserRepositoryForAcceptInvite struct{ mock.Mock }

func (m *mockAdminUserRepositoryForAcceptInvite) FindByEmail(ctx context.Context, email string) (*domain.AdminUser, error) {
	args := m.Called(ctx, email)
	if v := args.Get(0); v != nil {
		return v.(*domain.AdminUser), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockAdminUserRepositoryForAcceptInvite) FindByID(ctx context.Context, id string) (*domain.AdminUser, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*domain.AdminUser), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockAdminUserRepositoryForAcceptInvite) Save(ctx context.Context, user *domain.AdminUser) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockAdminUserRepositoryForAcceptInvite) Update(ctx context.Context, user *domain.AdminUser) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockAdminUserRepositoryForAcceptInvite) UpdatePassword(ctx context.Context, adminID, hash string) error {
	return m.Called(ctx, adminID, hash).Error(0)
}
func (m *mockAdminUserRepositoryForAcceptInvite) UpdateStatus(ctx context.Context, adminID, status string) error {
	return m.Called(ctx, adminID, status).Error(0)
}
func (m *mockAdminUserRepositoryForAcceptInvite) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockAdminUserRepositoryForAcceptInvite) GetRoleCodeByID(ctx context.Context, roleID string) (string, error) {
	args := m.Called(ctx, roleID)
	return args.String(0), args.Error(1)
}
func (m *mockAdminUserRepositoryForAcceptInvite) ActivateAccount(ctx context.Context, adminID, newHash string) error {
	return m.Called(ctx, adminID, newHash).Error(0)
}
func (m *mockAdminUserRepositoryForAcceptInvite) FindByEmailOrSSO(ctx context.Context, email string, ssoProvider string, ssoID string) (*domain.AdminUser, error) {
	args := m.Called(ctx, email, ssoProvider, ssoID)
	if v := args.Get(0); v != nil {
		return v.(*domain.AdminUser), args.Error(1)
	}
	return nil, args.Error(1)
}


// ---------------------------------------------------------------------------
// Mock: AuditLogger (accept-invite specific; reuses struct name scoped to file)
// ---------------------------------------------------------------------------

type mockAcceptInviteAuditLogger struct{ mock.Mock }

func (m *mockAcceptInviteAuditLogger) LogLoginSuccess(ctx context.Context, userID, email string) {
	m.Called(ctx, userID, email)
}
func (m *mockAcceptInviteAuditLogger) LogLoginFailed(ctx context.Context, email, reason string) {
	m.Called(ctx, email, reason)
}
func (m *mockAcceptInviteAuditLogger) LogLogoutSuccess(ctx context.Context, token string) {
	m.Called(ctx, token)
}
func (m *mockAcceptInviteAuditLogger) LogSessionRefreshed(ctx context.Context, userID string) {
	m.Called(ctx, userID)
}
func (m *mockAcceptInviteAuditLogger) LogAccountLocked(ctx context.Context, email string) {
	m.Called(ctx, email)
}
func (m *mockAcceptInviteAuditLogger) LogMFAChallengeIssued(ctx context.Context, userID, email string) {
	m.Called(ctx, userID, email)
}
func (m *mockAcceptInviteAuditLogger) LogMFASuccess(ctx context.Context, userID, email string) {
	m.Called(ctx, userID, email)
}
func (m *mockAcceptInviteAuditLogger) LogMFAFailed(ctx context.Context, email, reason string) {
	m.Called(ctx, email, reason)
}
func (m *mockAcceptInviteAuditLogger) LogPasswordResetRequested(ctx context.Context, event events.PasswordResetRequestedEvent) {
	m.Called(ctx, event)
}
func (m *mockAcceptInviteAuditLogger) LogPasswordResetCompleted(ctx context.Context, event events.PasswordResetCompletedEvent) {
	m.Called(ctx, event)
}
func (m *mockAcceptInviteAuditLogger) LogInviteAccepted(ctx context.Context, event events.InviteAcceptedEvent) {
	m.Called(ctx, event)
}
func (m *mockAcceptInviteAuditLogger) LogAdminActivated(ctx context.Context, event events.AdminActivatedEvent) {
	m.Called(ctx, event)
}
func (m *mockAcceptInviteAuditLogger) LogSSOSuccess(ctx context.Context, event events.SSOSuccessEvent) {
	m.Called(ctx, event)
}
func (m *mockAcceptInviteAuditLogger) LogSSOFailed(ctx context.Context, event events.SSOFailedEvent) {
	m.Called(ctx, event)
}

// ---------------------------------------------------------------------------
// Mock: NotificationPublisher
// ---------------------------------------------------------------------------

type mockNotificationPublisher struct{ mock.Mock }

func (m *mockNotificationPublisher) PublishInviteAcceptedNotification(ctx context.Context, event events.NotificationSendEvent) error {
	return m.Called(ctx, event).Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const validPassword = "Str0ng!Pass#1"

func newValidInviteToken() *domain.InviteToken {
	return &domain.InviteToken{
		ID:        "token-id-001",
		AdminID:   "admin-uuid-001",
		Token:     "raw-secure-token",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		CreatedBy: "inviter-uuid-001",
		CreatedAt: time.Now().UTC(),
	}
}

func newAcceptInviteUseCase(
	userRepo *mockAdminUserRepositoryForAcceptInvite,
	tokenRepo *mockInviteTokenRepository,
	audit *mockAcceptInviteAuditLogger,
	pub *mockNotificationPublisher,
) *AcceptInviteUseCase {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewAcceptInviteUseCase(userRepo, tokenRepo, audit, pub, logger)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAcceptInviteUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	inviteToken := newValidInviteToken()
	adminUser := &domain.AdminUser{ID: inviteToken.AdminID, Email: "admin@hros.io"}

	userRepo := new(mockAdminUserRepositoryForAcceptInvite)
	tokenRepo := new(mockInviteTokenRepository)
	audit := new(mockAcceptInviteAuditLogger)
	pub := new(mockNotificationPublisher)

	tokenRepo.On("FindByToken", ctx, inviteToken.Token).Return(inviteToken, nil)
	userRepo.On("ActivateAccount", ctx, inviteToken.AdminID, mock.MatchedBy(func(hash string) bool {
		// Verify the stored value is a bcrypt hash of validPassword.
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(validPassword)) == nil
	})).Return(nil)
	tokenRepo.On("Consume", ctx, inviteToken.Token).Return(inviteToken, nil)
	userRepo.On("FindByID", ctx, inviteToken.AdminID).Return(adminUser, nil)
	audit.On("LogInviteAccepted", ctx, mock.MatchedBy(func(e events.InviteAcceptedEvent) bool {
		return e.AdminID == inviteToken.AdminID &&
			e.InviteTokenID == inviteToken.ID &&
			e.InvitedBy == inviteToken.CreatedBy &&
			e.Email == adminUser.Email
	})).Once()
	audit.On("LogAdminActivated", ctx, mock.MatchedBy(func(e events.AdminActivatedEvent) bool {
		return e.AdminID == inviteToken.AdminID && e.Email == adminUser.Email
	})).Once()
	pub.On("PublishInviteAcceptedNotification", ctx, mock.MatchedBy(func(e events.NotificationSendEvent) bool {
		return e.RecipientID == inviteToken.CreatedBy && e.Type == "invite.accepted"
	})).Return(nil)

	uc := newAcceptInviteUseCase(userRepo, tokenRepo, audit, pub)
	err := uc.Execute(ctx, AcceptInviteInput{
		Token:     inviteToken.Token,
		Password:  validPassword,
		IPAddress: "127.0.0.1",
		UserAgent: "Go-test",
	})

	require.NoError(t, err)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
	pub.AssertExpectations(t)
}

func TestAcceptInviteUseCase_Execute_EmptyToken(t *testing.T) {
	uc := newAcceptInviteUseCase(nil, nil, nil, nil)
	err := uc.Execute(context.Background(), AcceptInviteInput{Token: "", Password: validPassword})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invite token cannot be empty")
}

func TestAcceptInviteUseCase_Execute_EmptyPassword(t *testing.T) {
	uc := newAcceptInviteUseCase(nil, nil, nil, nil)
	err := uc.Execute(context.Background(), AcceptInviteInput{Token: "tok", Password: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestAcceptInviteUseCase_Execute_WeakPassword(t *testing.T) {
	cases := []struct {
		name     string
		password string
	}{
		{"too short", "Short1!"},
		{"no uppercase", "alllower1!ab"},
		{"no digit", "NoDigitHere!!"},
		{"no special", "NoSpecial1234"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc := newAcceptInviteUseCase(nil, nil, nil, nil)
			err := uc.Execute(context.Background(), AcceptInviteInput{Token: "tok", Password: tc.password})
			require.ErrorIs(t, err, domainErrors.ErrPasswordWeak)
		})
	}
}

func TestAcceptInviteUseCase_Execute_FindByTokenError(t *testing.T) {
	ctx := context.Background()
	tokenRepo := new(mockInviteTokenRepository)
	tokenRepo.On("FindByToken", ctx, "tok").Return(nil, errors.New("db error"))

	uc := newAcceptInviteUseCase(nil, tokenRepo, nil, nil)
	err := uc.Execute(ctx, AcceptInviteInput{Token: "tok", Password: validPassword})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find invite token")
}

func TestAcceptInviteUseCase_Execute_TokenExpired(t *testing.T) {
	ctx := context.Background()
	expiredToken := newValidInviteToken()
	expiredToken.ExpiresAt = time.Now().Add(-1 * time.Hour) // expired

	tokenRepo := new(mockInviteTokenRepository)
	tokenRepo.On("FindByToken", ctx, expiredToken.Token).Return(expiredToken, nil)

	uc := newAcceptInviteUseCase(nil, tokenRepo, nil, nil)
	err := uc.Execute(ctx, AcceptInviteInput{Token: expiredToken.Token, Password: validPassword})
	require.ErrorIs(t, err, domainErrors.ErrInviteExpired)
}

func TestAcceptInviteUseCase_Execute_TokenAlreadyUsed(t *testing.T) {
	ctx := context.Background()
	usedToken := newValidInviteToken()
	now := time.Now()
	usedToken.UsedAt = &now

	tokenRepo := new(mockInviteTokenRepository)
	tokenRepo.On("FindByToken", ctx, usedToken.Token).Return(usedToken, nil)

	uc := newAcceptInviteUseCase(nil, tokenRepo, nil, nil)
	err := uc.Execute(ctx, AcceptInviteInput{Token: usedToken.Token, Password: validPassword})
	require.ErrorIs(t, err, domainErrors.ErrInviteUsed)
}

func TestAcceptInviteUseCase_Execute_ActivateAccountError(t *testing.T) {
	ctx := context.Background()
	inviteToken := newValidInviteToken()

	userRepo := new(mockAdminUserRepositoryForAcceptInvite)
	tokenRepo := new(mockInviteTokenRepository)
	tokenRepo.On("FindByToken", ctx, inviteToken.Token).Return(inviteToken, nil)
	userRepo.On("ActivateAccount", ctx, inviteToken.AdminID, mock.AnythingOfType("string")).Return(errors.New("db error"))

	uc := newAcceptInviteUseCase(userRepo, tokenRepo, nil, nil)
	err := uc.Execute(ctx, AcceptInviteInput{Token: inviteToken.Token, Password: validPassword})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "activate account")
}

func TestAcceptInviteUseCase_Execute_ConsumeTokenError(t *testing.T) {
	ctx := context.Background()
	inviteToken := newValidInviteToken()

	userRepo := new(mockAdminUserRepositoryForAcceptInvite)
	tokenRepo := new(mockInviteTokenRepository)
	tokenRepo.On("FindByToken", ctx, inviteToken.Token).Return(inviteToken, nil)
	userRepo.On("ActivateAccount", ctx, inviteToken.AdminID, mock.AnythingOfType("string")).Return(nil)
	tokenRepo.On("Consume", ctx, inviteToken.Token).Return(nil, errors.New("consume error"))

	uc := newAcceptInviteUseCase(userRepo, tokenRepo, nil, nil)
	err := uc.Execute(ctx, AcceptInviteInput{Token: inviteToken.Token, Password: validPassword})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "consume invite token")
}

func TestAcceptInviteUseCase_Execute_PublishNotificationError(t *testing.T) {
	// Since notification publish is now best-effort, a Kafka failure must NOT
	// surface as an error — the activation has already been committed.
	ctx := context.Background()
	inviteToken := newValidInviteToken()
	adminUser := &domain.AdminUser{ID: inviteToken.AdminID, Email: "admin@hros.io"}

	userRepo := new(mockAdminUserRepositoryForAcceptInvite)
	tokenRepo := new(mockInviteTokenRepository)
	audit := new(mockAcceptInviteAuditLogger)
	pub := new(mockNotificationPublisher)

	tokenRepo.On("FindByToken", ctx, inviteToken.Token).Return(inviteToken, nil)
	userRepo.On("ActivateAccount", ctx, inviteToken.AdminID, mock.AnythingOfType("string")).Return(nil)
	tokenRepo.On("Consume", ctx, inviteToken.Token).Return(inviteToken, nil)
	userRepo.On("FindByID", ctx, inviteToken.AdminID).Return(adminUser, nil)
	audit.On("LogInviteAccepted", ctx, mock.Anything).Once()
	audit.On("LogAdminActivated", ctx, mock.Anything).Once()
	pub.On("PublishInviteAcceptedNotification", ctx, mock.Anything).Return(errors.New("kafka error"))

	uc := newAcceptInviteUseCase(userRepo, tokenRepo, audit, pub)
	err := uc.Execute(ctx, AcceptInviteInput{Token: inviteToken.Token, Password: validPassword})

	// Notification failure is best-effort: Execute must succeed.
	require.NoError(t, err)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
	pub.AssertExpectations(t)
}
