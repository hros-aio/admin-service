package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hros/admin-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTTokenProvider(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	issuer := "test-issuer"
	p := NewJWTTokenProvider(privateKey, issuer)
	ctx := context.Background()

	t.Run("GenerateAccessToken", func(t *testing.T) {
		user := &domain.AdminUser{
			ID:    "user-123",
			Email: "admin@example.com",
			RoleID: "role-admin",
		}
		expiry := 15 * time.Minute

		tokenStr, err := p.GenerateAccessToken(ctx, user, expiry)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		// Parse and verify
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return &privateKey.PublicKey, nil
		})
		assert.NoError(t, err)
		assert.True(t, token.Valid)

		claims, ok := token.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, issuer, claims["iss"])
		assert.Equal(t, user.ID, claims["sub"])
		assert.Equal(t, user.Email, claims["email"])
		assert.Equal(t, user.RoleID, claims["role"])
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		token, err := p.GenerateRefreshToken(ctx)
		assert.NoError(t, err)
		assert.Len(t, token, 64) // 32 bytes hex encoded
	})
}
