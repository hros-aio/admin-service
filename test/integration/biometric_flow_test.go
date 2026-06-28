// Package integration provides integration tests for the HROS Admin Service.
package integration

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/alicebob/miniredis/v2"
	adapterHttp "github.com/hros/admin-service/internal/adapter/http"
	"github.com/hros/admin-service/internal/adapter/http/auth/dto"
	kafkaProducer "github.com/hros/admin-service/internal/adapter/kafka/producer"
	"github.com/hros/admin-service/internal/application"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/config"
	"github.com/hros/admin-service/internal/domain"
	authInfra "github.com/hros/admin-service/internal/infrastructure/auth"
	authCache "github.com/hros/admin-service/internal/infrastructure/cache"
	authRepo "github.com/hros/admin-service/internal/infrastructure/repository/auth"
	"github.com/hros/admin-service/internal/platform/database"
	httpPlatform "github.com/hros/admin-service/internal/platform/http"
	"github.com/hros/admin-service/internal/platform/redis"
	sharedErrors "github.com/hros/admin-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Test helper: ECDSA key pair generation
// ---------------------------------------------------------------------------

// generateTestECDSAKeyPair creates a P-256 ECDSA key pair.
// Returns the private key and the public key as a base64-std-encoded DER PKIX
// bytes string (ready for storage in the webauthn_credentials JSONB column).
func generateTestECDSAKeyPair(t *testing.T) (*ecdsa.PrivateKey, string) {
	t.Helper()
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "generate ECDSA P-256 key")

	pubKeyDER, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err, "marshal PKIX public key")

	pubKeyB64 := base64.StdEncoding.EncodeToString(pubKeyDER)
	return privKey, pubKeyB64
}

// ---------------------------------------------------------------------------
// Test helper: authenticatorData construction
// ---------------------------------------------------------------------------

// buildAuthenticatorData constructs a minimal 37-byte WebAuthn authenticatorData buffer.
//
//	Bytes  0-31: SHA-256("localhost")  — RP ID hash (matches usecase allowlist)
//	Byte   32:   0x05                  — UP (bit 0) and UV (bit 2) flags set
//	Bytes 33-36: signCount as big-endian uint32
func buildAuthenticatorData(signCount uint32) []byte {
	rpIDHash := sha256.Sum256([]byte("localhost"))
	authData := make([]byte, 37)
	copy(authData[:32], rpIDHash[:])
	authData[32] = 0x05 // UP | UV
	binary.BigEndian.PutUint32(authData[33:37], signCount)
	return authData
}

// ---------------------------------------------------------------------------
// Test helper: clientDataJSON construction
// ---------------------------------------------------------------------------

// buildClientDataJSON returns the raw JSON bytes for a WebAuthn assertion
// clientDataJSON using the challenge extracted from the challenge response.
// Origin is fixed to "http://localhost:3000" (usecase allowlist).
func buildClientDataJSON(challengeB64url string) []byte {
	cd := struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Origin    string `json:"origin"`
	}{
		Type:      "webauthn.get",
		Challenge: challengeB64url,
		Origin:    "http://localhost:3000",
	}
	b, _ := json.Marshal(cd)
	return b
}

// ---------------------------------------------------------------------------
// Test helper: ECDSA assertion signing
// ---------------------------------------------------------------------------

// signWebAuthnAssertion computes the ECDSA P-256 signature over the WebAuthn
// signed data as required by verifyECDSASignature in verify_biometric_usecase.go:
//
//	signedDataHash = SHA-256( authenticatorData || SHA-256(clientDataJSON) )
//
// The signature is ASN.1 DER-encoded (R, S integers).
func signWebAuthnAssertion(t *testing.T, privKey *ecdsa.PrivateKey, clientDataJSON, authData []byte) []byte {
	t.Helper()
	clientDataHash := sha256.Sum256(clientDataJSON)
	signedData := append(authData, clientDataHash[:]...)
	signedDataHash := sha256.Sum256(signedData)

	r, s, err := ecdsa.Sign(rand.Reader, privKey, signedDataHash[:])
	require.NoError(t, err, "ECDSA sign")

	type ecdsaSig struct {
		R, S *big.Int
	}
	sigDER, err := asn1.Marshal(ecdsaSig{R: r, S: s})
	require.NoError(t, err, "ASN.1 marshal signature")
	return sigDER
}

// ---------------------------------------------------------------------------
// Test helper: base64url encoding
// ---------------------------------------------------------------------------

// b64url returns the base64 RawURL (no padding) encoding of b.
func b64url(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// Main integration test
// ---------------------------------------------------------------------------

func TestBiometricFlow(t *testing.T) {
	ctx := context.Background()

	// T001-T006: helpers defined above.

	// -------------------------------------------------------------------------
	// T007: PostgreSQL testcontainer + miniredis setup
	// -------------------------------------------------------------------------
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("hros_admin"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, postgresContainer.Terminate(ctx))
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// Run all required migrations using the dollar-quote-aware helper.
	// 000003 adds webauthn_credentials column and contains $$...$$  PL/pgSQL blocks.
	migDir := findMigrationsDir(t)
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// -------------------------------------------------------------------------
	// T008: Seed test data
	// -------------------------------------------------------------------------

	// Generate ECDSA P-256 key pair for test credential
	privKey, pubKeyB64 := generateTestECDSAKeyPair(t)

	const (
		testEmail    = "biometric-test@hros.com"
		testCredID   = "test-cred-id"
		testPassword = "Password123!"
	)

	roleID := domain.NewUUID()
	err = db.Exec(
		"INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		roleID, "Standard Admin", "Standard Admin Role", false,
	).Error
	require.NoError(t, err)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), 12)
	require.NoError(t, err)

	adminUserID := domain.NewUUID()
	err = db.Exec(
		"INSERT INTO admin_users (id, name, email, password_hash, role_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		adminUserID, "Biometric Test Admin", testEmail, string(hashedPassword), roleID, "active",
	).Error
	require.NoError(t, err)

	// Seed webauthn_credentials JSONB with sign_count = 0
	credJSON := fmt.Sprintf(`{"id":%q,"public_key":%q,"sign_count":0}`, testCredID, pubKeyB64)
	err = db.Exec(
		"UPDATE admin_users SET webauthn_credentials = ?::jsonb WHERE email = ?",
		credJSON, testEmail,
	).Error
	require.NoError(t, err)

	// -------------------------------------------------------------------------
	// T009: Bootstrap Fx application (mirror auth_flow_test.go)
	// -------------------------------------------------------------------------
	testPort := getFreePort(t)

	opts := fx.Options(
		fx.NopLogger,
		fx.Provide(func() (*config.Config, error) {
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				return nil, err
			}
			privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
			pemBlock := &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: privBytes,
			}
			pemBytes := pem.EncodeToMemory(pemBlock)
			return &config.Config{
				AppName:       "admin-service-biometric-integration-test",
				Env:           "test",
				Port:          testPort,
				DBURL:         connStr,
				RedisURL:      "redis://" + mr.Addr(),
				KafkaBrokers:  []string{"localhost:9092"},
				LogLevel:      "debug",
				JWTPrivateKey: string(pemBytes),
			}, nil
		}),
		fx.Provide(func() *slog.Logger {
			return slog.New(slog.NewTextHandler(io.Discard, nil))
		}),
		fx.Provide(database.NewDatabase),
		fx.Provide(database.NewTxManager),
		fx.Provide(authRepo.NewGormAdminUserRepository),
		fx.Provide(authRepo.NewGormSessionTokenRepository),
		fx.Provide(redis.NewRedisClient),
		fx.Provide(authCache.NewRedisTokenBlacklist),
		fx.Provide(authCache.NewRedisBruteForceCache),
		fx.Provide(authCache.NewRedisMFACache),
		fx.Provide(authCache.NewRedisWebAuthnChallengeCache),
		fx.Provide(authCache.NewRedisPasswordResetCache),
		fx.Provide(func() (sarama.SyncProducer, error) {
			return mocks.NewSyncProducer(t, nil), nil
		}),
		fx.Provide(func() (sarama.ConsumerGroup, error) {
			return &mockConsumerGroup{}, nil
		}),
		authInfra.Module,
		application.Module,
		kafkaProducer.Module,
		fx.Provide(func(p *kafkaProducer.EmailKafkaProducer) interfaces.LockoutNotifier { return p }),
		fx.Provide(func(p *kafkaProducer.EmailKafkaProducer) interfaces.PasswordResetNotifier { return p }),
		fx.Provide(httpPlatform.NewHealthHandler),
		fx.Provide(httpPlatform.NewServer),
		adapterHttp.Module,
	)

	app := fx.New(opts)
	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startCancel()
	require.NoError(t, app.Start(startCtx))
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		assert.NoError(t, app.Stop(stopCtx))
	}()

	baseURL := fmt.Sprintf("http://localhost:%d", testPort)
	healthClient := &http.Client{Timeout: 1 * time.Second}
	biometricClient := &http.Client{Timeout: 15 * time.Second}

	// Poll /health until ready
	var ready bool
	for i := 0; i < 50; i++ {
		resp, err := healthClient.Get(baseURL + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.True(t, ready, "server failed to become ready in time")

	// -------------------------------------------------------------------------
	// T010: HappyPath_ChallengeAndVerify
	// -------------------------------------------------------------------------
	t.Run("HappyPath_ChallengeAndVerify", func(t *testing.T) {
		// Step 1 — POST /v1/auth/biometric/challenge
		challengeReqBody, _ := json.Marshal(dto.BiometricChallengeRequest{Email: testEmail})
		challengeResp, err := biometricClient.Post(
			baseURL+"/v1/auth/biometric/challenge",
			"application/json",
			bytes.NewBuffer(challengeReqBody),
		)
		require.NoError(t, err)
		defer func() { _ = challengeResp.Body.Close() }()

		assert.Equal(t, http.StatusOK, challengeResp.StatusCode)

		var challengeOut dto.BiometricChallengeResponse
		require.NoError(t, json.NewDecoder(challengeResp.Body).Decode(&challengeOut))
		assert.NotEmpty(t, challengeOut.Challenge, "challenge must be non-empty")
		assert.Equal(t, testCredID, challengeOut.CredentialID)

		// Step 2 — Build cryptographic proof
		authData := buildAuthenticatorData(1) // sign_count = 1 (incremented from DB 0)
		clientDataJSON := buildClientDataJSON(challengeOut.Challenge)
		signature := signWebAuthnAssertion(t, privKey, clientDataJSON, authData)

		// Step 3 — POST /v1/auth/biometric/verify
		verifyReq := dto.BiometricVerifyRequest{
			Email:             testEmail,
			CredentialID:      testCredID,
			AuthenticatorData: b64url(authData),
			ClientDataJSON:    b64url(clientDataJSON),
			Signature:         b64url(signature),
			RememberMe:        false,
		}
		verifyReqBody, _ := json.Marshal(verifyReq)
		verifyResp, err := biometricClient.Post(
			baseURL+"/v1/auth/biometric/verify",
			"application/json",
			bytes.NewBuffer(verifyReqBody),
		)
		require.NoError(t, err)
		defer func() { _ = verifyResp.Body.Close() }()

		assert.Equal(t, http.StatusOK, verifyResp.StatusCode)

		var loginOut dto.LoginResponse
		require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&loginOut))
		assert.NotEmpty(t, loginOut.AccessToken, "access_token must be issued")
		assert.NotEmpty(t, loginOut.RefreshToken, "refresh_token must be issued")

		// Step 4 — Assert sign_count was incremented in the database
		var rawCredsStr string
		err = db.Raw(
			"SELECT webauthn_credentials FROM admin_users WHERE email = ?", testEmail,
		).Scan(&rawCredsStr).Error
		require.NoError(t, err)

		var cred struct {
			SignCount uint32 `json:"sign_count"`
		}
		require.NoError(t, json.Unmarshal([]byte(rawCredsStr), &cred))
		assert.Equal(t, uint32(1), cred.SignCount, "sign_count must be incremented to 1 after verify")
	})

	// -------------------------------------------------------------------------
	// T011: InvalidSignature_Returns401
	// -------------------------------------------------------------------------
	t.Run("InvalidSignature_Returns401", func(t *testing.T) {
		// Issue a fresh challenge
		challengeReqBody, _ := json.Marshal(dto.BiometricChallengeRequest{Email: testEmail})
		challengeResp, err := biometricClient.Post(
			baseURL+"/v1/auth/biometric/challenge",
			"application/json",
			bytes.NewBuffer(challengeReqBody),
		)
		require.NoError(t, err)
		defer func() { _ = challengeResp.Body.Close() }()
		require.Equal(t, http.StatusOK, challengeResp.StatusCode)

		var challengeOut dto.BiometricChallengeResponse
		require.NoError(t, json.NewDecoder(challengeResp.Body).Decode(&challengeOut))

		// Build valid authenticatorData and clientDataJSON but corrupt the signature
		authData := buildAuthenticatorData(2)
		clientDataJSON := buildClientDataJSON(challengeOut.Challenge)
		validSig := signWebAuthnAssertion(t, privKey, clientDataJSON, authData)

		// Flip the first byte to corrupt the signature
		corruptedSig := make([]byte, len(validSig))
		copy(corruptedSig, validSig)
		corruptedSig[0] ^= 0xFF

		verifyReq := dto.BiometricVerifyRequest{
			Email:             testEmail,
			CredentialID:      testCredID,
			AuthenticatorData: b64url(authData),
			ClientDataJSON:    b64url(clientDataJSON),
			Signature:         b64url(corruptedSig),
			RememberMe:        false,
		}
		verifyReqBody, _ := json.Marshal(verifyReq)
		verifyResp, err := biometricClient.Post(
			baseURL+"/v1/auth/biometric/verify",
			"application/json",
			bytes.NewBuffer(verifyReqBody),
		)
		require.NoError(t, err)
		defer func() { _ = verifyResp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, verifyResp.StatusCode)

		var errResp sharedErrors.ErrorResponse
		require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&errResp))
		assert.Equal(t, "unauthorized", errResp.Code)
	})

	// -------------------------------------------------------------------------
	// T012: ExpiredChallenge_Returns401
	// -------------------------------------------------------------------------
	t.Run("ExpiredChallenge_Returns401", func(t *testing.T) {
		// Build a well-formed verify request without issuing a challenge first.
		// Use a synthetic random challenge so the Redis key does not exist.
		syntheticChallenge := make([]byte, 32)
		_, err := rand.Read(syntheticChallenge)
		require.NoError(t, err)

		syntheticChallengeB64 := b64url(syntheticChallenge)
		authData := buildAuthenticatorData(3)
		clientDataJSON := buildClientDataJSON(syntheticChallengeB64)
		signature := signWebAuthnAssertion(t, privKey, clientDataJSON, authData)

		verifyReq := dto.BiometricVerifyRequest{
			Email:             testEmail,
			CredentialID:      testCredID,
			AuthenticatorData: b64url(authData),
			ClientDataJSON:    b64url(clientDataJSON),
			Signature:         b64url(signature),
			RememberMe:        false,
		}
		verifyReqBody, _ := json.Marshal(verifyReq)
		verifyResp, err := biometricClient.Post(
			baseURL+"/v1/auth/biometric/verify",
			"application/json",
			bytes.NewBuffer(verifyReqBody),
		)
		require.NoError(t, err)
		defer func() { _ = verifyResp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, verifyResp.StatusCode)

		var errResp sharedErrors.ErrorResponse
		require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&errResp))
		assert.Equal(t, "unauthorized", errResp.Code)
	})
}
