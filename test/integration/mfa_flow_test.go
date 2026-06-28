// Package integration provides integration tests for the HROS Admin Service.
package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/golang-jwt/jwt/v5"
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
	"github.com/pquerna/otp/totp"
	goRedis "github.com/redis/go-redis/v9"
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

// knownTotpSecret is a well-known Base32 TOTP secret used for seeding test data.
// It matches the format required by github.com/pquerna/otp/totp for code generation.
const knownTotpSecret = "JBSWY3DPEHPK3PXP"

// invalidTOTPCode generates a valid TOTP code from the given secret and then
// deterministically corrupts it by incrementing the last digit, guaranteeing
// the returned code is never coincidentally valid.
func invalidTOTPCode(t *testing.T, secret string) string {
	t.Helper()
	valid, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err, "TOTP code generation must succeed")
	// Flip the last digit by cycling it: '9'→'0', others +1.
	runes := []rune(valid)
	last := runes[len(runes)-1]
	if last == '9' {
		runes[len(runes)-1] = '0'
	} else {
		runes[len(runes)-1] = last + 1
	}
	return string(runes)
}

// TestSuperAdminMFALoginFlow tests the full end-to-end Super Admin login flow
// with MFA enforcement using real PostgreSQL and Redis containers.
//
// Flow:
//  1. Seed a Super Admin user with a known TOTP secret.
//  2. POST /v1/auth/login → expect MFA challenge (mfa_required=true, mfa_token set).
//  3. Assert the mfa_token key exists in Redis with the expected 5-minute TTL.
//  4. Generate a valid TOTP code from the known secret.
//  5. POST /v1/auth/mfa/verify → expect 200 OK with structurally valid JWTs.
//
// Definition of Done:
//   - The intermediate 5-minute Redis state mapping is proven (mfa_token exists in
//     Redis with TTL ≤ 300 s between login and verify).
//   - Correct JWT generation is proven upon MFA success (tokens parse and verify
//     successfully against the RSA public key kept in test scope).
func TestSuperAdminMFALoginFlow(t *testing.T) {
	ctx := context.Background()

	// ── 1. Start a PostgreSQL testcontainer ───────────────────────────────────
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

	// Connect directly with GORM to run migrations and seed test data.
	// FIX (line 89-90): retrieve the underlying *sql.DB and defer Close() so the
	// connection pool is cleaned up at the end of the test lifecycle.
	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()

	// ── 2. Run database migrations (all three) ────────────────────────────────
	migDir := findMigrationsDir(t)
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))

	// ── 3. Seed a Super Admin role ────────────────────────────────────────────
	superAdminRoleID := domain.NewUUID()
	err = db.Exec(
		"INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		superAdminRoleID, "Super Admin", "Super Admin Role", true,
	).Error
	require.NoError(t, err)

	// ── 4. Seed a Super Admin user with a known TOTP secret ───────────────────
	// The totp_secret column holds the Base32 TOTP secret used for code generation.
	superAdminPassword := "SuperSecret123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(superAdminPassword), 12)
	require.NoError(t, err)

	superAdminUserID := domain.NewUUID()
	superAdminEmail := "super-admin-mfa@hros.com"
	err = db.Exec(
		`INSERT INTO admin_users (id, name, email, password_hash, role_id, status, mfa_enabled, totp_secret)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		superAdminUserID,
		"MFA Super Admin",
		superAdminEmail,
		string(hashedPassword),
		superAdminRoleID,
		"active",
		true,
		knownTotpSecret,
	).Error
	require.NoError(t, err)

	// ── 5. Start a Redis testcontainer ────────────────────────────────────────
	redisContainer, redisURL, err := runRedisContainer(ctx)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, redisContainer.Terminate(ctx))
	}()

	// ── 6. Generate the RSA key at test scope so the public key remains ────────
	// available for JWT structural validation after the Fx app is started.
	// FIX (line 306-307): the private key is generated here rather than inside
	// the Fx provider closure so the public key can be used to verify tokens.
	testPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privBytes := x509.MarshalPKCS1PrivateKey(testPrivateKey)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	// ── 7. Bootstrap the Fx application with the test infrastructure ──────────
	testPort := getFreePort(t)

	opts := fx.Options(
		fx.NopLogger,
		fx.Provide(func() (*config.Config, error) {
			return &config.Config{
				AppName:       "admin-service-mfa-flow-test",
				Env:           "test",
				Port:          testPort,
				DBURL:         connStr,
				RedisURL:      redisURL,
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

	startCtx, startCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startCancel()

	err = app.Start(startCtx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		assert.NoError(t, app.Stop(stopCtx))
	}()

	baseURL := fmt.Sprintf("http://localhost:%d", testPort)
	healthClient := &http.Client{Timeout: 1 * time.Second}
	authClient := &http.Client{Timeout: 15 * time.Second}

	// Poll /health until the server is ready.
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

	// Build a direct Redis client against the same container so subtests can
	// inspect cache state independently of the application layer.
	// FIX (line 61-63): used to assert the mfa_token Redis key and its TTL.
	redisOpts, err := goRedis.ParseURL(redisURL)
	require.NoError(t, err)
	testRedisClient := goRedis.NewClient(redisOpts)
	defer func() { _ = testRedisClient.Close() }()

	// ── 8. POST /v1/auth/login — expect MFA challenge ─────────────────────────
	t.Run("login returns MFA challenge for Super Admin", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    superAdminEmail,
			Password: superAdminPassword,
		}
		loginBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		resp, err := authClient.Post(
			baseURL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(loginBody),
		)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// The login response must be 200 OK with mfa_required=true and an mfa_token.
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp dto.LoginResponse
		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err)

		assert.True(t, loginResp.MFARequired, "expected mfa_required to be true for Super Admin login")
		assert.NotEmpty(t, loginResp.MFAToken, "expected mfa_token to be non-empty")
		assert.Empty(t, loginResp.AccessToken, "expected access_token to be absent during MFA challenge")
		assert.Empty(t, loginResp.RefreshToken, "expected refresh_token to be absent during MFA challenge")
		assert.NotEmpty(t, loginResp.MFAMethods, "expected mfa_methods to be present")
		assert.Contains(t, loginResp.MFAMethods, "totp", "expected totp in mfa_methods")

		// FIX (line 61-63): assert the mfa_token key exists in Redis with the
		// expected ≤5-minute lifetime, directly proving the intermediate state.
		if loginResp.MFAToken != "" {
			redisKey := "auth:mfa_token:" + loginResp.MFAToken
			ttl, err := testRedisClient.TTL(ctx, redisKey).Result()
			require.NoError(t, err, "Redis TTL lookup must succeed")
			assert.Positive(t, ttl, "mfa_token Redis key must exist after login")
			assert.LessOrEqual(t, ttl, 5*time.Minute,
				"mfa_token TTL must be at most 5 minutes")
		}
	})

	// ── 9. POST /v1/auth/mfa/verify — expect JWT tokens on success ─────────────
	t.Run("mfa verify completes login and returns JWT tokens", func(t *testing.T) {
		// Step A: Re-issue a new login to get a fresh mfa_token (subtests are independent).
		loginReq := dto.LoginRequest{
			Email:    superAdminEmail,
			Password: superAdminPassword,
		}
		loginBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		loginResp, err := authClient.Post(
			baseURL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(loginBody),
		)
		require.NoError(t, err)
		defer func() { _ = loginResp.Body.Close() }()
		require.Equal(t, http.StatusOK, loginResp.StatusCode)

		var mfaChallenge dto.LoginResponse
		require.NoError(t, json.NewDecoder(loginResp.Body).Decode(&mfaChallenge))
		require.True(t, mfaChallenge.MFARequired)
		require.NotEmpty(t, mfaChallenge.MFAToken)

		// FIX (line 61-63): prove the intermediate Redis state exists with a TTL ≤ 5 min.
		redisKey := "auth:mfa_token:" + mfaChallenge.MFAToken
		ttl, err := testRedisClient.TTL(ctx, redisKey).Result()
		require.NoError(t, err, "Redis TTL lookup must succeed")
		assert.Positive(t, ttl, "mfa_token Redis key must exist between login and verify")
		assert.LessOrEqual(t, ttl, 5*time.Minute, "mfa_token TTL must be at most 5 minutes")

		// Step B: Generate a valid TOTP code from the known secret at this moment.
		totpCode, err := totp.GenerateCode(knownTotpSecret, time.Now())
		require.NoError(t, err, "TOTP code generation must succeed with known secret")

		// Step C: POST /v1/auth/mfa/verify with the mfa_token and generated code.
		verifyReq := dto.MFAVerifyRequest{
			MFAToken: mfaChallenge.MFAToken,
			Method:   "totp",
			Code:     totpCode,
		}
		verifyBody, err := json.Marshal(verifyReq)
		require.NoError(t, err)

		verifyResp, err := authClient.Post(
			baseURL+"/v1/auth/mfa/verify",
			"application/json",
			bytes.NewBuffer(verifyBody),
		)
		require.NoError(t, err)
		defer func() { _ = verifyResp.Body.Close() }()

		// Step D: Assert that the response is 200 OK with structurally valid JWTs.
		assert.Equal(t, http.StatusOK, verifyResp.StatusCode)

		var tokenResp dto.LoginResponse
		err = json.NewDecoder(verifyResp.Body).Decode(&tokenResp)
		require.NoError(t, err)

		require.NotEmpty(t, tokenResp.AccessToken, "expected non-empty access_token after MFA verification")
		require.NotEmpty(t, tokenResp.RefreshToken, "expected non-empty refresh_token after MFA verification")
		assert.False(t, tokenResp.MFARequired, "expected mfa_required to be false in success response")
		assert.Empty(t, tokenResp.MFAToken, "expected mfa_token to be absent in success response")

		// FIX (line 306-307): validate the access token is a structurally valid
		// RS256 JWT signed by the key we generated at test scope.
		accessToken, err := jwt.Parse(tokenResp.AccessToken, func(tok *jwt.Token) (interface{}, error) {
			if _, ok := tok.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
			}
			return &testPrivateKey.PublicKey, nil
		})
		require.NoError(t, err, "access_token must parse and verify against the RSA public key")
		assert.True(t, accessToken.Valid, "access_token must be a valid JWT")

		claims, ok := accessToken.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.NotEmpty(t, claims["sub"], "access_token must carry a sub claim")
		assert.NotEmpty(t, claims["jti"], "access_token must carry a jti claim")
		assert.Equal(t, "hros-admin", claims["iss"], "access_token issuer must be hros-admin")

		// The refresh token is a 64-character hex string (32 random bytes), not a JWT.
		assert.Len(t, tokenResp.RefreshToken, 64,
			"refresh_token must be a 64-character hex-encoded random string")
	})

	// ── 10. Expired/consumed mfa_token is rejected ──────────────────────────────
	t.Run("replayed mfa_token is rejected after use", func(t *testing.T) {
		// Obtain a fresh mfa_token.
		loginReq := dto.LoginRequest{
			Email:    superAdminEmail,
			Password: superAdminPassword,
		}
		loginBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		loginResp, err := authClient.Post(
			baseURL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(loginBody),
		)
		require.NoError(t, err)
		defer func() { _ = loginResp.Body.Close() }()
		require.Equal(t, http.StatusOK, loginResp.StatusCode)

		var mfaChallenge dto.LoginResponse
		require.NoError(t, json.NewDecoder(loginResp.Body).Decode(&mfaChallenge))
		require.NotEmpty(t, mfaChallenge.MFAToken)

		capturedToken := mfaChallenge.MFAToken

		// First verify — must succeed and delete the token from Redis.
		totpCode, err := totp.GenerateCode(knownTotpSecret, time.Now())
		require.NoError(t, err)

		firstVerifyReq := dto.MFAVerifyRequest{
			MFAToken: capturedToken,
			Method:   "totp",
			Code:     totpCode,
		}
		firstBody, err := json.Marshal(firstVerifyReq)
		require.NoError(t, err)

		firstResp, err := authClient.Post(
			baseURL+"/v1/auth/mfa/verify",
			"application/json",
			bytes.NewBuffer(firstBody),
		)
		require.NoError(t, err)
		defer func() { _ = firstResp.Body.Close() }()
		require.Equal(t, http.StatusOK, firstResp.StatusCode,
			"first MFA verify must succeed")

		// Second verify with the same token — must be rejected as expired/consumed.
		// Generate a new TOTP code (the token itself is what is deleted from Redis).
		secondTotpCode, err := totp.GenerateCode(knownTotpSecret, time.Now())
		require.NoError(t, err)

		secondVerifyReq := dto.MFAVerifyRequest{
			MFAToken: capturedToken, // same token, now deleted from Redis
			Method:   "totp",
			Code:     secondTotpCode,
		}
		secondBody, err := json.Marshal(secondVerifyReq)
		require.NoError(t, err)

		secondResp, err := authClient.Post(
			baseURL+"/v1/auth/mfa/verify",
			"application/json",
			bytes.NewBuffer(secondBody),
		)
		require.NoError(t, err)
		defer func() { _ = secondResp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, secondResp.StatusCode,
			"replayed mfa_token must be rejected")

		var errResp sharedErrors.ErrorResponse
		err = json.NewDecoder(secondResp.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Equal(t, "MFA_TOKEN_EXPIRED", errResp.Code,
			"replayed token must return MFA_TOKEN_EXPIRED error code")
	})

	// ── 11. Invalid TOTP code is rejected ──────────────────────────────────────
	t.Run("invalid TOTP code is rejected with MFA_INVALID", func(t *testing.T) {
		loginReq := dto.LoginRequest{
			Email:    superAdminEmail,
			Password: superAdminPassword,
		}
		loginBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		loginResp, err := authClient.Post(
			baseURL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(loginBody),
		)
		require.NoError(t, err)
		defer func() { _ = loginResp.Body.Close() }()
		require.Equal(t, http.StatusOK, loginResp.StatusCode)

		var mfaChallenge dto.LoginResponse
		require.NoError(t, json.NewDecoder(loginResp.Body).Decode(&mfaChallenge))
		require.NotEmpty(t, mfaChallenge.MFAToken)

		// FIX (line 412-416): generate a deterministically invalid code by
		// deriving a valid TOTP and mutating the last digit, so the code is
		// guaranteed to be wrong regardless of the current TOTP window.
		badCode := invalidTOTPCode(t, knownTotpSecret)

		verifyReq := dto.MFAVerifyRequest{
			MFAToken: mfaChallenge.MFAToken,
			Method:   "totp",
			Code:     badCode,
		}
		verifyBody, err := json.Marshal(verifyReq)
		require.NoError(t, err)

		verifyResp, err := authClient.Post(
			baseURL+"/v1/auth/mfa/verify",
			"application/json",
			bytes.NewBuffer(verifyBody),
		)
		require.NoError(t, err)
		defer func() { _ = verifyResp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, verifyResp.StatusCode)

		var errResp sharedErrors.ErrorResponse
		err = json.NewDecoder(verifyResp.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Equal(t, "MFA_INVALID", errResp.Code,
			"invalid TOTP code must return MFA_INVALID error code")
	})

	// ── 12. Standard Admin login bypasses MFA ──────────────────────────────────
	t.Run("standard admin login does not trigger MFA challenge", func(t *testing.T) {
		// Seed a Standard Admin role and user.
		standardRoleID := domain.NewUUID()
		err := db.Exec(
			"INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
			standardRoleID, "Standard Admin", "Standard Admin Role", false,
		).Error
		require.NoError(t, err)

		standardPassword := "StandardSecret123!"
		hashedStdPwd, err := bcrypt.GenerateFromPassword([]byte(standardPassword), 12)
		require.NoError(t, err)

		standardUserID := domain.NewUUID()
		standardEmail := "standard-admin-mfa-flow@hros.com"
		err = db.Exec(
			`INSERT INTO admin_users (id, name, email, password_hash, role_id, status)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			standardUserID,
			"Standard Admin User",
			standardEmail,
			string(hashedStdPwd),
			standardRoleID,
			"active",
		).Error
		require.NoError(t, err)

		loginReq := dto.LoginRequest{
			Email:    standardEmail,
			Password: standardPassword,
		}
		loginBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		resp, err := authClient.Post(
			baseURL+"/v1/auth/login",
			"application/json",
			bytes.NewBuffer(loginBody),
		)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp dto.LoginResponse
		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err)

		assert.False(t, loginResp.MFARequired, "standard admin must not be challenged for MFA")
		assert.Empty(t, loginResp.MFAToken, "standard admin must not receive mfa_token")
		assert.NotEmpty(t, loginResp.AccessToken, "standard admin must receive access_token directly")
		assert.NotEmpty(t, loginResp.RefreshToken, "standard admin must receive refresh_token directly")
	})
}
