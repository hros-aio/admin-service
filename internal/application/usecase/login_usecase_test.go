package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type mockUserRepo struct{ mock.Mock }
func (m *mockUserRepo) Save(ctx context.Context, u *domain.AdminUser) error { return m.Called(ctx, u).Error(0) }
func (m *mockUserRepo) Update(ctx context.Context, u *domain.AdminUser) error { return m.Called(ctx, u).Error(0) }
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.AdminUser, error) { 
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*domain.AdminUser), args.Error(1)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.AdminUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*domain.AdminUser), args.Error(1)
}
func (m *mockUserRepo) Delete(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }

type mockSessionRepo struct{ mock.Mock }
func (m *mockSessionRepo) Save(ctx context.Context, t *domain.SessionToken) error { return m.Called(ctx, t).Error(0) }
func (m *mockSessionRepo) FindByToken(ctx context.Context, t string) (*domain.SessionToken, error) { return nil, nil }
func (m *mockSessionRepo) DeleteByToken(ctx context.Context, t string) error { return m.Called(ctx, t).Error(0) }
func (m *mockSessionRepo) DeleteByAdminID(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }
func (m *mockSessionRepo) Revoke(ctx context.Context, t string, r string) error { return m.Called(ctx, t, r).Error(0) }

type mockPasswordHelper struct{ mock.Mock }
func (m *mockPasswordHelper) Hash(p string) (string, error) { return "", nil }
func (m *mockPasswordHelper) Compare(h, p string) error { return m.Called(h, p).Error(0) }
func (m *mockPasswordHelper) CompareDummy(p string) { m.Called(p) }

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
func (m *mockAuditLogger) LogLoginSuccess(ctx context.Context, id, email string) { m.Called(ctx, id, email) }
func (m *mockAuditLogger) LogLoginFailed(ctx context.Context, email, reason string) { m.Called(ctx, email, reason) }
func (m *mockAuditLogger) LogLogoutSuccess(ctx context.Context, token string) { m.Called(ctx, token) }

func TestLoginUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	sessionRepo := new(mockSessionRepo)
	password := new(mockPasswordHelper)
	tokens := new(mockTokenProvider)
	audit := new(mockAuditLogger)

	uc := NewLoginUseCase(userRepo, sessionRepo, password, tokens, audit)

	input := LoginInput{
		Email: "admin@example.com",
		Password: "password123",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}

	t.Run("Success", func(t *testing.T) {
		user := &domain.AdminUser{
			ID: "user-id",
			Email: input.Email,
			PasswordHash: "hashed",
			Status: domain.AdminUserStatusActive,
		}

		userRepo.On("FindByEmail", ctx, input.Email).Return(user, nil).Once()
		password.On("Compare", "hashed", input.Password).Return(nil).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("refresh-token", nil).Once()
		sessionRepo.On("Save", ctx, mock.AnythingOfType("*domain.SessionToken")).Return(nil).Once()
		audit.On("LogLoginSuccess", ctx, user.ID, user.Email).Return().Once()

		output, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "access-token", output.AccessToken)
		assert.Equal(t, "refresh-token", output.RefreshToken)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		tokens.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userRepo.On("FindByEmail", ctx, input.Email).Return(nil, domainErrors.ErrUserNotFound).Once()
		password.On("CompareDummy", input.Password).Return().Once()
		audit.On("LogLoginFailed", ctx, input.Email, "user not found").Return().Once()

		output, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		user := &domain.AdminUser{
			ID: "user-id",
			Email: input.Email,
			PasswordHash: "hashed",
			Status: domain.AdminUserStatusActive,
		}

		userRepo.On("FindByEmail", ctx, input.Email).Return(user, nil).Once()
		password.On("Compare", "hashed", input.Password).Return(errors.New("invalid")).Once()
		audit.On("LogLoginFailed", ctx, input.Email, "invalid password").Return().Once()

		output, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		userRepo.AssertExpectations(t)
		password.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("UserLocked", func(t *testing.T) {
		lockedUntil := time.Now().Add(1 * time.Hour)
		user := &domain.AdminUser{
			ID: "user-id",
			Email: input.Email,
			Status: domain.AdminUserStatusActive,
			LockedUntil: &lockedUntil,
		}

		userRepo.On("FindByEmail", ctx, input.Email).Return(user, nil).Once()
		password.On("CompareDummy", input.Password).Return().Once()
		audit.On("LogLoginFailed", ctx, input.Email, "user locked").Return().Once()

		output, err := uc.Execute(ctx, input)

		assert.ErrorIs(t, err, domainErrors.ErrUserLocked)
		assert.Nil(t, output)
	})
}
