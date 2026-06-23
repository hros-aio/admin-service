package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVerifyMFAUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	mfaToken := "valid-mfa-token"
	adminID := "admin-id-123"
	email := "admin@example.com"
	totpSecret := "JBSWY3DPEHPK3PXP" // Valid Base32 secret

	t.Run("successfully validates TOTP and issues tokens", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		// Generate a valid code for right now
		code, err := totp.GenerateCode(totpSecret, time.Now())
		assert.NoError(t, err)

		user := &domain.AdminUser{
			ID:         adminID,
			Email:      email,
			Name:       "Admin User",
			TotpSecret: totpSecret,
			Status:     "active",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()
		audit.On("LogMFASuccess", ctx, adminID, email).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("refresh-token", nil).Once()
		sessionRepo.On("Save", ctx, mock.MatchedBy(func(s *domain.SessionToken) bool {
			return s.AdminID == adminID && s.RefreshToken == "refresh-token" && s.IsPersistent == true
		})).Return(nil).Once()
		mfaCache.On("DeleteToken", ctx, mfaToken).Return(nil).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		input := VerifyMFAInput{
			MFAToken:   mfaToken,
			Method:     "totp",
			Code:       code,
			RememberMe: true,
			IPAddress:  "127.0.0.1",
			UserAgent:  "Mozilla",
		}
		out, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, "access-token", out.AccessToken)
		assert.Equal(t, "refresh-token", out.RefreshToken)
		assert.Equal(t, adminID, out.User.ID)
		assert.Equal(t, email, out.User.Email)

		mfaCache.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		audit.AssertExpectations(t)
		tokens.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("returns ErrMFATokenExpired if cache lookup returns token-expired error", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		mfaCache.On("GetAdminID", ctx, mfaToken).Return("", domainErrors.ErrMFATokenExpired).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: "123456"})

		assert.Nil(t, out)
		assert.ErrorIs(t, err, domainErrors.ErrMFATokenExpired)
	})

	t.Run("returns ErrMFATokenExpired if cache lookup returns empty ID", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		mfaCache.On("GetAdminID", ctx, mfaToken).Return("", nil).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: "123456"})

		assert.Nil(t, out)
		assert.ErrorIs(t, err, domainErrors.ErrMFATokenExpired)
	})

	t.Run("fails if cache lookup fails with general error", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		mfaCache.On("GetAdminID", ctx, mfaToken).Return("", errors.New("redis down")).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: "123456"})

		assert.Nil(t, out)
		assert.ErrorContains(t, err, "failed to retrieve admin ID from cache")
	})

	t.Run("fails if user lookup fails", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(nil, errors.New("db error")).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: "123456"})

		assert.Nil(t, out)
		assert.ErrorContains(t, err, "failed to find user by ID")
	})

	t.Run("fails if user is inactive", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		user := &domain.AdminUser{
			ID:     adminID,
			Status: "inactive",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: "123456"})

		assert.Nil(t, out)
		assert.ErrorIs(t, err, domainErrors.ErrUserInactive)
	})

	t.Run("fails if user is locked", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		futureTime := time.Now().Add(10 * time.Minute)
		user := &domain.AdminUser{
			ID:          adminID,
			Status:      "active",
			LockedUntil: &futureTime,
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: "123456"})

		assert.Nil(t, out)
		assert.ErrorIs(t, err, domainErrors.ErrUserLocked)
	})

	t.Run("fails if user TotpSecret is empty", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		user := &domain.AdminUser{
			ID:     adminID,
			Email:  email,
			Status: "active",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()
		audit.On("LogMFAFailed", ctx, email, "TOTP is not configured").Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: "123456"})

		assert.Nil(t, out)
		assert.ErrorIs(t, err, domainErrors.ErrMFAInvalid)
	})

	t.Run("fails on incorrect TOTP code", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		user := &domain.AdminUser{
			ID:         adminID,
			Email:      email,
			TotpSecret: totpSecret,
			Status:     "active",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()
		audit.On("LogMFAFailed", ctx, email, "invalid TOTP code").Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: "000000"}) // Invalid code

		assert.Nil(t, out)
		assert.ErrorIs(t, err, domainErrors.ErrMFAInvalid)
	})

	t.Run("fails on unsupported verification method", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		user := &domain.AdminUser{
			ID:         adminID,
			Email:      email,
			TotpSecret: totpSecret,
			Status:     "active",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()
		audit.On("LogMFAFailed", ctx, email, "unsupported method: webauthn").Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "webauthn", Code: "123456"})

		assert.Nil(t, out)
		assert.ErrorContains(t, err, "unsupported verification method")
	})

	t.Run("fails if generate access token fails", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		code, err := totp.GenerateCode(totpSecret, time.Now())
		assert.NoError(t, err)

		user := &domain.AdminUser{
			ID:         adminID,
			Email:      email,
			TotpSecret: totpSecret,
			Status:     "active",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()
		audit.On("LogMFASuccess", ctx, adminID, email).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("", errors.New("token error")).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: code})

		assert.Nil(t, out)
		assert.ErrorContains(t, err, "generate access token")
	})

	t.Run("fails if generate refresh token fails", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		code, err := totp.GenerateCode(totpSecret, time.Now())
		assert.NoError(t, err)

		user := &domain.AdminUser{
			ID:         adminID,
			Email:      email,
			TotpSecret: totpSecret,
			Status:     "active",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()
		audit.On("LogMFASuccess", ctx, adminID, email).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("", errors.New("refresh error")).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: code})

		assert.Nil(t, out)
		assert.ErrorContains(t, err, "generate refresh token")
	})

	t.Run("fails if save session fails", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		code, err := totp.GenerateCode(totpSecret, time.Now())
		assert.NoError(t, err)

		user := &domain.AdminUser{
			ID:         adminID,
			Email:      email,
			TotpSecret: totpSecret,
			Status:     "active",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()
		audit.On("LogMFASuccess", ctx, adminID, email).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("refresh-token", nil).Once()
		sessionRepo.On("Save", ctx, mock.Anything).Return(errors.New("db error")).Once()

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: code})

		assert.Nil(t, out)
		assert.ErrorContains(t, err, "save session token")
	})

	t.Run("does not fail verification if cache token deletion fails", func(t *testing.T) {
		userRepo := new(mockUserRepo)
		sessionRepo := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)
		mfaCache := new(mockMFACache)

		code, err := totp.GenerateCode(totpSecret, time.Now())
		assert.NoError(t, err)

		user := &domain.AdminUser{
			ID:         adminID,
			Email:      email,
			TotpSecret: totpSecret,
			Status:     "active",
		}

		mfaCache.On("GetAdminID", ctx, mfaToken).Return(adminID, nil).Once()
		userRepo.On("FindByID", ctx, adminID).Return(user, nil).Once()
		audit.On("LogMFASuccess", ctx, adminID, email).Once()
		tokens.On("GenerateAccessToken", ctx, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", ctx).Return("refresh-token", nil).Once()
		sessionRepo.On("Save", ctx, mock.Anything).Return(nil).Once()
		mfaCache.On("DeleteToken", ctx, mfaToken).Return(errors.New("redis err")).Once() // cache delete fails

		uc := NewVerifyMFAUseCase(userRepo, sessionRepo, tokens, audit, mfaCache, slog.Default())
		out, err := uc.Execute(ctx, VerifyMFAInput{MFAToken: mfaToken, Method: "totp", Code: code})

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, "access-token", out.AccessToken)
		assert.Equal(t, "refresh-token", out.RefreshToken)
	})
}
