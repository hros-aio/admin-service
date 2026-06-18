package auth

import (
	"context"
	"crypto/rsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hros/admin-service/internal/domain"
)

// JWTTokenProvider implements the TokenProvider interface using RS256 JWTs.
type JWTTokenProvider struct {
	privateKey *rsa.PrivateKey
	issuer     string
}

// NewJWTTokenProvider creates a new JWTTokenProvider.
func NewJWTTokenProvider(privateKey *rsa.PrivateKey, issuer string) *JWTTokenProvider {
	return &JWTTokenProvider{
		privateKey: privateKey,
		issuer:     issuer,
	}
}

// GenerateAccessToken generates an RS256 JWT access token.
func (p *JWTTokenProvider) GenerateAccessToken(ctx context.Context, user *domain.AdminUser, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": p.issuer,
		"sub": user.ID,
		"exp": now.Add(expiry).Unix(),
		"iat": now.Unix(),
		"email": user.Email,
		"role": user.RoleID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(p.privateKey)
}

// GenerateRefreshToken generates a secure random refresh token.
func (p *JWTTokenProvider) GenerateRefreshToken(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
