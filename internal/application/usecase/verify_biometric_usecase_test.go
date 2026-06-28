package usecase

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"github.com/hros/admin-service/internal/domain"
	domainErrors "github.com/hros/admin-service/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func generateTestKeyPair(t *testing.T) (*ecdsa.PrivateKey, string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ecdsa key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})
	return priv, string(pemBytes)
}

func signTestAssertion(priv *ecdsa.PrivateKey, clientDataJSON, authenticatorData []byte) []byte {
	clientDataHash := sha256.Sum256(clientDataJSON)
	signedData := append(authenticatorData, clientDataHash[:]...)
	signedDataHash := sha256.Sum256(signedData)

	r, s, err := ecdsa.Sign(rand.Reader, priv, signedDataHash[:])
	if err != nil {
		panic(err)
	}

	sig := ecdsaSignature{R: r, S: s}
	sigBytes, err := asn1.Marshal(sig)
	if err != nil {
		panic(err)
	}
	return sigBytes
}

func makeTestAuthenticatorData(counter uint32) []byte {
	// W3C Authenticator Data structure has signature counter at bytes 33-36.
	authData := make([]byte, 37)
	binary.BigEndian.PutUint32(authData[33:37], counter)
	return authData
}

func TestVerifyBiometricUseCase(t *testing.T) {
	email := "admin@example.com"
	credentialID := "cred_123"
	cachedChallenge := []byte("secure_challenge_bytes")
	encodedChallenge := base64.RawURLEncoding.EncodeToString(cachedChallenge)

	clientDataJSON, _ := json.Marshal(map[string]string{
		"challenge": encodedChallenge,
	})

	privKey, pubKeyPEM := generateTestKeyPair(t)
	authData := makeTestAuthenticatorData(10) // authenticator count = 10
	validSignature := signTestAssertion(privKey, clientDataJSON, authData)

	t.Run("success verification and session generation", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		sr := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)

		uc := NewVerifyBiometricUseCase(ur, cache, sr, tokens, audit)

		credentialsJSON, _ := json.Marshal(WebAuthnCredential{
			ID:        credentialID,
			PublicKey: pubKeyPEM,
			SignCount: 5, // stored count = 5
		})

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			Name:                "John Admin",
			Status:              "active",
			WebauthnCredentials: credentialsJSON,
		}

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()
		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()
		ur.On("UpdateWebAuthnSignCount", mock.Anything, user.ID, uint32(10)).Return(nil).Once()
		audit.On("LogBiometricSuccess", mock.Anything, mock.Anything).Return().Once()

		tokens.On("GenerateAccessToken", mock.Anything, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", mock.Anything).Return("refresh-token", nil).Once()
		sr.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authData,
			Signature:         validSignature,
			RememberMe:        false,
			IPAddress:         "127.0.0.1",
			UserAgent:         "Mozilla/5.0",
		})

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, "access-token", out.AccessToken)
		assert.Equal(t, "refresh-token", out.RefreshToken)
		assert.Equal(t, user.ID, out.User.ID)
		assert.Equal(t, user.Email, out.User.Email)

		ur.AssertExpectations(t)
		cache.AssertExpectations(t)
		sr.AssertExpectations(t)
		tokens.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("missing required parameters error", func(t *testing.T) {
		uc := NewVerifyBiometricUseCase(nil, nil, nil, nil, nil)
		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email: "",
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.Contains(t, err.Error(), "missing required verification parameters")
	})

	t.Run("challenge expired or not found", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(([]byte)(nil), domainErrors.ErrChallengeNotFoundOrExpired).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authData,
			Signature:         validSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrChallengeNotFoundOrExpired)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
	})

	t.Run("malformed clientDataJSON returns invalid signature error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    []byte(`{"malformed"`),
			AuthenticatorData: authData,
			Signature:         validSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidBiometricSignature)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
	})

	t.Run("challenge mismatch returns invalid signature error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		mismatchedClientData, _ := json.Marshal(map[string]string{
			"challenge": "wrong_challenge",
		})

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    mismatchedClientData,
			AuthenticatorData: authData,
			Signature:         validSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidBiometricSignature)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
	})

	t.Run("user not found returns biometric not registered", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()
		ur.On("FindByEmail", mock.Anything, email).Return((*domain.AdminUser)(nil), domainErrors.ErrUserNotFound).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authData,
			Signature:         validSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrBiometricNotRegistered)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
		ur.AssertExpectations(t)
	})

	t.Run("user is inactive returns error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		user := &domain.AdminUser{
			ID:     "user-uuid",
			Email:  email,
			Status: "pending",
		}

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()
		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authData,
			Signature:         validSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrUserInactive)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
		ur.AssertExpectations(t)
	})

	t.Run("user is locked returns error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		lockedTime := time.Now().Add(10 * time.Minute)
		user := &domain.AdminUser{
			ID:          "user-uuid",
			Email:       email,
			Status:      "active",
			LockedUntil: &lockedTime,
		}

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()
		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authData,
			Signature:         validSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrUserLocked)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
		ur.AssertExpectations(t)
	})

	t.Run("no registered biometric credential error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			Status:              "active",
			WebauthnCredentials: nil,
		}

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()
		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authData,
			Signature:         validSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrBiometricNotRegistered)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
		ur.AssertExpectations(t)
	})

	t.Run("cryptographic signature mismatch returns invalid signature error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		credentialsJSON, _ := json.Marshal(WebAuthnCredential{
			ID:        credentialID,
			PublicKey: pubKeyPEM,
			SignCount: 5,
		})

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			Status:              "active",
			WebauthnCredentials: credentialsJSON,
		}

		badSignature := []byte("bad_signature_bytes")

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()
		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authData,
			Signature:         badSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidBiometricSignature)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
		ur.AssertExpectations(t)
	})

	t.Run("cloning detected (sign count <= stored sign count) error", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		uc := NewVerifyBiometricUseCase(ur, cache, nil, nil, nil)

		credentialsJSON, _ := json.Marshal(WebAuthnCredential{
			ID:        credentialID,
			PublicKey: pubKeyPEM,
			SignCount: 15, // stored count = 15
		})

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			Status:              "active",
			WebauthnCredentials: credentialsJSON,
		}

		authDataCloned := makeTestAuthenticatorData(15) // authenticator returns 15 (<= 15)
		clonedSignature := signTestAssertion(privKey, clientDataJSON, authDataCloned)

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()
		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authDataCloned,
			Signature:         clonedSignature,
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidBiometricSignature)
		assert.Nil(t, out)
		cache.AssertExpectations(t)
		ur.AssertExpectations(t)
	})

	t.Run("session save error propagates", func(t *testing.T) {
		ur := new(mockUserRepo)
		cache := new(mockChallengeCache)
		sr := new(mockSessionRepo)
		tokens := new(mockTokenProvider)
		audit := new(mockAuditLogger)

		uc := NewVerifyBiometricUseCase(ur, cache, sr, tokens, audit)

		credentialsJSON, _ := json.Marshal(WebAuthnCredential{
			ID:        credentialID,
			PublicKey: pubKeyPEM,
			SignCount: 5,
		})

		user := &domain.AdminUser{
			ID:                  "user-uuid",
			Email:               email,
			Status:              "active",
			WebauthnCredentials: credentialsJSON,
		}

		dbErr := errors.New("db save failure")

		cache.On("VerifyAndConsumeChallenge", mock.Anything, email).Return(cachedChallenge, nil).Once()
		ur.On("FindByEmail", mock.Anything, email).Return(user, nil).Once()
		ur.On("UpdateWebAuthnSignCount", mock.Anything, user.ID, uint32(10)).Return(nil).Once()
		audit.On("LogBiometricSuccess", mock.Anything, mock.Anything).Return().Once()

		tokens.On("GenerateAccessToken", mock.Anything, user, 15*time.Minute).Return("access-token", nil).Once()
		tokens.On("GenerateRefreshToken", mock.Anything).Return("refresh-token", nil).Once()
		sr.On("Save", mock.Anything, mock.Anything).Return(dbErr).Once()

		out, err := uc.Execute(context.Background(), VerifyBiometricInput{
			Email:             email,
			CredentialID:      credentialID,
			ClientDataJSON:    clientDataJSON,
			AuthenticatorData: authData,
			Signature:         validSignature,
		})

		assert.ErrorIs(t, err, dbErr)
		assert.Nil(t, out)

		ur.AssertExpectations(t)
		cache.AssertExpectations(t)
		sr.AssertExpectations(t)
		tokens.AssertExpectations(t)
		audit.AssertExpectations(t)
	})
}
