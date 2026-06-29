package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
)

// LogoutInput represents the input for the logout use case.
type LogoutInput struct {
	RefreshToken string
	AccessToken  string
}

// LogoutUseCase handles the admin logout process by revoking/deleting the session token
// and blacklisting the current access token.
type LogoutUseCase struct {
	sessionRepo domain.SessionTokenRepository
	blacklist   interfaces.TokenBlacklist
	audit       authDomain.AuditLogger
}

// NewLogoutUseCase creates a new LogoutUseCase.
func NewLogoutUseCase(
	sessionRepo domain.SessionTokenRepository,
	blacklist interfaces.TokenBlacklist,
	audit authDomain.AuditLogger,
) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
		blacklist:   blacklist,
		audit:       audit,
	}
}

// Execute performs the logout flow.
func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	if input.RefreshToken == "" {
		return errors.New("refresh token cannot be empty")
	}

	// 1. Extract JTI and exp from AccessToken (if provided) and blacklist it
	if input.AccessToken != "" {
		var claims jwt.MapClaims
		token, _, err := new(jwt.Parser).ParseUnverified(input.AccessToken, &claims)
		if err == nil && token != nil {
			var jti string
			if jtiVal, ok := claims["jti"].(string); ok {
				jti = jtiVal
			}

			var exp float64
			if expVal, ok := claims["exp"].(float64); ok {
				exp = expVal
			}

			if jti != "" && exp > 0 {
				expiryTime := time.Unix(int64(exp), 0)
				ttl := time.Until(expiryTime)
				if ttl > 0 {
					if err := uc.blacklist.Add(ctx, jti, ttl); err != nil {
						slog.WarnContext(
							ctx, "failed to blacklist access token JTI during logout",
							slog.String("jti", jti),
							slog.Any("error", err),
						)
					}
				}
			}
		} else if err != nil {
			slog.WarnContext(
				ctx, "failed to parse access token during logout",
				slog.Any("error", err),
			)
		}
	}

	// 2. Delete the token from session_tokens via repository.
	// Since GORM Delete is idempotent and returns nil if the record is not found,
	// this works for both existing and non-existing tokens.
	if err := uc.sessionRepo.DeleteByToken(ctx, input.RefreshToken); err != nil {
		return fmt.Errorf("failed to delete session token: %w", err)
	}

	// 3. Emit logout.success to audit log interface
	uc.audit.LogLogoutSuccess(ctx, input.RefreshToken)

	return nil
}
