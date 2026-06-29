package usecase

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/hros/admin-service/internal/application/auth"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/domain"
	authDomain "github.com/hros/admin-service/internal/domain/auth"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/hros/admin-service/internal/domain/events"
)

// VerifyBiometricInput represents the parameters required to execute biometric login verification.
type VerifyBiometricInput struct {
	Email             string
	CredentialID      string
	ClientDataJSON    []byte
	AuthenticatorData []byte
	Signature         []byte
	RememberMe        bool
	IPAddress         string
	UserAgent         string
}

// VerifyBiometricUseCase orchestrates WebAuthn signature verification and session issuance.
type VerifyBiometricUseCase struct {
	userRepo       domain.AdminUserRepository
	challengeCache interfaces.WebAuthnChallengeCache
	sessionRepo    domain.SessionTokenRepository
	tokens         auth.TokenProvider
	audit          authDomain.AuditLogger
}

// NewVerifyBiometricUseCase creates a new VerifyBiometricUseCase.
func NewVerifyBiometricUseCase(
	userRepo domain.AdminUserRepository,
	challengeCache interfaces.WebAuthnChallengeCache,
	sessionRepo domain.SessionTokenRepository,
	tokens auth.TokenProvider,
	audit authDomain.AuditLogger,
) *VerifyBiometricUseCase {
	return &VerifyBiometricUseCase{
		userRepo:       userRepo,
		challengeCache: challengeCache,
		sessionRepo:    sessionRepo,
		tokens:         tokens,
		audit:          audit,
	}
}

type clientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

type ecdsaSignature struct {
	R, S *big.Int
}

// Execute performs WebAuthn cryptographic assertion check, updates the sign count monotonically, and returns user session.
func (uc *VerifyBiometricUseCase) Execute(ctx context.Context, input VerifyBiometricInput) (*LoginOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if email == "" || input.CredentialID == "" || len(input.ClientDataJSON) == 0 || len(input.AuthenticatorData) == 0 || len(input.Signature) == 0 {
		return nil, errors.New("missing required verification parameters")
	}

	// 1. Retrieve and consume the cached challenge from Redis
	cachedChallenge, err := uc.challengeCache.VerifyAndConsumeChallenge(ctx, email)
	if err != nil {
		if errors.Is(err, domainErrors.ErrChallengeNotFoundOrExpired) {
			return nil, domainErrors.ErrChallengeNotFoundOrExpired
		}
		return nil, fmt.Errorf("verify and consume challenge: %w", err)
	}

	// 2. Parse clientDataJSON and verify type is webauthn.get, origin is allowed, and the challenge matches
	if err := verifyClientData(input.ClientDataJSON, cachedChallenge); err != nil {
		return nil, err
	}

	// 3. Retrieve user
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserNotFound) {
			return nil, domainErrors.ErrBiometricNotRegistered
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if !user.IsActive() {
		return nil, domainErrors.ErrUserInactive
	}

	if user.IsLocked() {
		return nil, domainErrors.ErrUserLocked
	}

	// Validate authenticator data length, RP ID hash, and UP/UV flags before credential parsing
	if err := verifyAuthenticatorData(input.AuthenticatorData); err != nil {
		return nil, err
	}

	// 4. Retrieve and parse WebAuthn Credentials
	creds, err := parseCredentials(user.WebauthnCredentials)
	if err != nil {
		return nil, err
	}

	var matchedCred *WebAuthnCredential
	for i := range creds {
		if creds[i].ID == input.CredentialID {
			matchedCred = &creds[i]
			break
		}
	}

	if matchedCred == nil || matchedCred.PublicKey == "" {
		return nil, domainErrors.ErrBiometricNotRegistered
	}

	// 5. Cryptographically verify the ECDSA signature
	if !verifyECDSASignature(matchedCred.PublicKey, input.ClientDataJSON, input.AuthenticatorData, input.Signature) {
		return nil, domainErrors.ErrInvalidBiometricSignature
	}

	// 6. Monotonicity check on signature count (cloning detection)
	newCount := binary.BigEndian.Uint32(input.AuthenticatorData[33:37])
	if (matchedCred.SignCount > 0 || newCount > 0) && newCount <= matchedCred.SignCount {
		return nil, domainErrors.ErrInvalidBiometricSignature
	}

	// 7. Update counter in DB (using monotonic repository helper) for matched credential
	if err := uc.userRepo.UpdateWebAuthnSignCount(ctx, user.ID, matchedCred.ID, newCount); err != nil {
		return nil, fmt.Errorf("failed to update sign count: %w", err)
	}

	// 8. Issue tokens and session
	accessToken, err := uc.tokens.GenerateAccessToken(ctx, user, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.tokens.GenerateRefreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

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
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// 9. Log biometric success audit trail ONLY after successful session persistence
	biometricEvent := events.BiometricSuccessEvent{
		AdminID:      user.ID,
		Email:        user.Email,
		CredentialID: matchedCred.ID,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		OccurredAt:   time.Now().UTC(),
	}
	uc.audit.LogBiometricSuccess(ctx, biometricEvent)

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

func verifyClientData(clientDataJSON []byte, cachedChallenge []byte) error {
	var cd clientData
	if err := json.Unmarshal(clientDataJSON, &cd); err != nil {
		return domainErrors.ErrInvalidBiometricSignature
	}

	if cd.Type != "webauthn.get" {
		return domainErrors.ErrInvalidBiometricSignature
	}

	allowedOrigins := map[string]bool{
		"http://localhost:3000":  true,
		"https://localhost:3000": true,
		"https://hros.admin":     true,
	}
	if !allowedOrigins[cd.Origin] {
		return domainErrors.ErrInvalidBiometricSignature
	}

	decodedChallenge, err := base64.RawURLEncoding.DecodeString(cd.Challenge)
	if err != nil || !bytes.Equal(cachedChallenge, decodedChallenge) {
		return domainErrors.ErrInvalidBiometricSignature
	}
	return nil
}

func verifyAuthenticatorData(authenticatorData []byte) error {
	if len(authenticatorData) < 37 {
		return domainErrors.ErrInvalidBiometricSignature
	}

	rpIDHash := authenticatorData[:32]
	expectedHash1 := sha256.Sum256([]byte("localhost"))
	expectedHash2 := sha256.Sum256([]byte("hros.admin"))
	if !bytes.Equal(rpIDHash, expectedHash1[:]) && !bytes.Equal(rpIDHash, expectedHash2[:]) {
		return domainErrors.ErrInvalidBiometricSignature
	}

	flags := authenticatorData[32]
	if (flags & 0x01) == 0 { // User Present (UP) must be set
		return domainErrors.ErrInvalidBiometricSignature
	}
	if (flags & 0x04) == 0 { // User Verified (UV) must be set
		return domainErrors.ErrInvalidBiometricSignature
	}
	return nil
}

func parseCredentials(webauthnCredentials []byte) ([]WebAuthnCredential, error) {
	trimmedCreds := bytes.TrimSpace(webauthnCredentials)
	if len(trimmedCreds) == 0 {
		return nil, domainErrors.ErrBiometricNotRegistered
	}

	var creds []WebAuthnCredential
	switch trimmedCreds[0] {
	case '[':
		if err := json.Unmarshal(trimmedCreds, &creds); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credentials array: %w", err)
		}
	case '{':
		var singleCred WebAuthnCredential
		if err := json.Unmarshal(trimmedCreds, &singleCred); err != nil {
			return nil, fmt.Errorf("failed to unmarshal single credential: %w", err)
		}
		creds = []WebAuthnCredential{singleCred}
	default:
		return nil, domainErrors.ErrBiometricNotRegistered
	}
	return creds, nil
}

func verifyECDSASignature(pubKeyStr string, clientDataJSON, authenticatorData, signature []byte) bool {
	var pubKeyBytes []byte
	block, _ := pem.Decode([]byte(pubKeyStr))
	if block != nil {
		pubKeyBytes = block.Bytes
	} else {
		decoded, err := base64.StdEncoding.DecodeString(pubKeyStr)
		if err == nil {
			pubKeyBytes = decoded
		} else {
			pubKeyBytes = []byte(pubKeyStr)
		}
	}

	pub, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return false
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok || ecdsaPub.Curve.Params().Name != "P-256" {
		return false
	}

	clientDataHash := sha256.Sum256(clientDataJSON)
	signedData := make([]byte, len(authenticatorData)+len(clientDataHash))
	copy(signedData, authenticatorData)
	copy(signedData[len(authenticatorData):], clientDataHash[:])
	signedDataHash := sha256.Sum256(signedData)

	var sig ecdsaSignature
	if _, err := asn1.Unmarshal(signature, &sig); err != nil {
		return false
	}

	return ecdsa.Verify(ecdsaPub, signedDataHash[:], sig.R, sig.S)
}
