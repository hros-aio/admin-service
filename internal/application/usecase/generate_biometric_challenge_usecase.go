package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
)

// GenerateBiometricChallengeInput represents the input payload for challenge generation.
type GenerateBiometricChallengeInput struct {
	Email string
}

// GenerateBiometricChallengeOutput represents the challenge output details returned to the client.
type GenerateBiometricChallengeOutput struct {
	Challenge    string
	CredentialID string
}

// WebAuthnCredential matches the structure stored inside the admin_users webauthn_credentials column.
type WebAuthnCredential struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	SignCount uint32 `json:"sign_count"`
}

// GenerateBiometricChallengeUseCase handles WebAuthn login challenge generation and caching.
type GenerateBiometricChallengeUseCase struct {
	userRepo       domain.AdminUserRepository
	challengeCache interfaces.WebAuthnChallengeCache
}

// NewGenerateBiometricChallengeUseCase creates a new GenerateBiometricChallengeUseCase.
func NewGenerateBiometricChallengeUseCase(
	userRepo domain.AdminUserRepository,
	challengeCache interfaces.WebAuthnChallengeCache,
) *GenerateBiometricChallengeUseCase {
	return &GenerateBiometricChallengeUseCase{
		userRepo:       userRepo,
		challengeCache: challengeCache,
	}
}

// Execute checks user eligibility, generates a cryptographically secure challenge, caches it, and returns the details.
func (uc *GenerateBiometricChallengeUseCase) Execute(ctx context.Context, input GenerateBiometricChallengeInput) (*GenerateBiometricChallengeOutput, error) {
	if input.Email == "" {
		return nil, errors.New("email cannot be empty")
	}

	// 1. Fetch user by email
	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserNotFound) {
			return nil, domainErrors.ErrBiometricNotRegistered
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 2. Validate webauthn_credentials column is not empty
	if len(user.WebauthnCredentials) == 0 {
		return nil, domainErrors.ErrBiometricNotRegistered
	}

	// 3. Parse credentials (JSONB) supporting both single-object and array schemas
	var creds []WebAuthnCredential
	firstChar := user.WebauthnCredentials[0]

	if firstChar == '[' {
		if err := json.Unmarshal(user.WebauthnCredentials, &creds); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credentials array: %w", err)
		}
	} else if firstChar == '{' {
		var singleCred WebAuthnCredential
		if err := json.Unmarshal(user.WebauthnCredentials, &singleCred); err != nil {
			return nil, fmt.Errorf("failed to unmarshal single credential: %w", err)
		}
		creds = []WebAuthnCredential{singleCred}
	} else {
		return nil, domainErrors.ErrBiometricNotRegistered
	}

	if len(creds) == 0 || creds[0].ID == "" {
		return nil, domainErrors.ErrBiometricNotRegistered
	}

	// 4. Generate 32-byte secure random challenge
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// 5. Store challenge in Redis with a 5-minute TTL
	ttl := 5 * time.Minute
	if err := uc.challengeCache.StoreChallenge(ctx, input.Email, challengeBytes, ttl); err != nil {
		return nil, fmt.Errorf("failed to store challenge in cache: %w", err)
	}

	// 6. Return raw base64url-encoded challenge and registered credential ID
	challengeStr := base64.RawURLEncoding.EncodeToString(challengeBytes)
	return &GenerateBiometricChallengeOutput{
		Challenge:    challengeStr,
		CredentialID: creds[0].ID,
	}, nil
}
