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
	"strings"
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

// runRedisContainer starts a generic Redis container using testcontainers-go.
func runRedisContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		_ = redisContainer.Terminate(ctx)
		return nil, "", err
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		_ = redisContainer.Terminate(ctx)
		return nil, "", err
	}

	redisURL := fmt.Sprintf("redis://%s:%s", host, port.Port())
	return redisContainer, redisURL, nil
}

// extractJTI unverified parses the token and returns the JTI claim value.
func extractJTI(tokenStr string) (string, error) {
	var claims jwt.MapClaims
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, &claims)
	if err != nil || token == nil {
		return "", fmt.Errorf("parse unverified: %w", err)
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return "", fmt.Errorf("jti claim not found or empty")
	}
	return jti, nil
}

func TestSessionPersistenceFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Setup testcontainers PostgreSQL instance
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
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Connect directly with GORM to run migrations and seed test data
	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// 2. Setup Redis testcontainer instance
	redisContainer, redisURL, err := runRedisContainer(ctx)
	require.NoError(t, err)
	defer func() {
		err := redisContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// 3. Execute database migrations
	migDir := findMigrationsDir(t)
	runSQLFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runSQLFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))

	// 4. Seed active Standard Admin user
	roleID := domain.NewUUID()
	err = db.Exec("INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		roleID, "Standard Admin", "Standard Admin Role", false).Error
	require.NoError(t, err)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	require.NoError(t, err)

	adminUserID := domain.NewUUID()
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		adminUserID, "Test Admin", "test-admin@hros.com", string(hashedPassword), roleID, "active").Error
	require.NoError(t, err)

	// 5. Bootstrap the Fx Echo application with custom overridden config
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
				AppName:       "admin-service-persistence-test",
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
	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startCancel()

	err = app.Start(startCtx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		err := app.Stop(stopCtx)
		assert.NoError(t, err)
	}()

	baseURL := fmt.Sprintf("http://localhost:%d", testPort)
	healthClient := &http.Client{Timeout: 1 * time.Second}
	authClient := &http.Client{Timeout: 15 * time.Second}

	// Poll the readiness endpoint (/health) to ensure the server has fully started.
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

	// 6. Execute POST /v1/auth/login with remember_me=true
	loginReq := dto.LoginRequest{
		Email:      "test-admin@hros.com",
		Password:   "password123",
		RememberMe: true,
	}
	loginReqBytes, err := json.Marshal(loginReq)
	require.NoError(t, err)
	resp, err := authClient.Post(baseURL+"/v1/auth/login", "application/json", bytes.NewBuffer(loginReqBytes))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp dto.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RefreshToken)

	// 7. Verify the 30-day DB session token persistence
	var session struct {
		RefreshToken string
		IsPersistent bool
		ExpiresAt    time.Time
	}
	err = db.Table("session_tokens").Where("refresh_token = ?", loginResp.RefreshToken).First(&session).Error
	require.NoError(t, err)
	assert.True(t, session.IsPersistent)
	expectedExpiry := time.Now().Add(30 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, session.ExpiresAt, 2*time.Minute)

	// 8. Perform session refresh rotation
	refreshReq := dto.RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	}
	refreshReqBytes, err := json.Marshal(refreshReq)
	require.NoError(t, err)
	refreshResp, err := authClient.Post(baseURL+"/v1/auth/refresh", "application/json", bytes.NewBuffer(refreshReqBytes))
	require.NoError(t, err)
	defer func() { _ = refreshResp.Body.Close() }()

	assert.Equal(t, http.StatusOK, refreshResp.StatusCode)

	var refreshResult dto.LoginResponse
	err = json.NewDecoder(refreshResp.Body).Decode(&refreshResult)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshResult.AccessToken)
	assert.NotEmpty(t, refreshResult.RefreshToken)
	assert.NotEqual(t, loginResp.RefreshToken, refreshResult.RefreshToken)

	// 9. Verify rotated session in DB still has 30 days persistent expiration
	var rotatedSession struct {
		RefreshToken string
		IsPersistent bool
		ExpiresAt    time.Time
	}
	err = db.Table("session_tokens").Where("refresh_token = ?", refreshResult.RefreshToken).First(&rotatedSession).Error
	require.NoError(t, err)
	assert.True(t, rotatedSession.IsPersistent)
	expectedRotatedExpiry := time.Now().Add(30 * 24 * time.Hour)
	assert.WithinDuration(t, expectedRotatedExpiry, rotatedSession.ExpiresAt, 2*time.Minute)

	// Verify old session token was deleted/updated
	var oldSessionCount int64
	err = db.Table("session_tokens").Where("refresh_token = ?", loginResp.RefreshToken).Count(&oldSessionCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), oldSessionCount)

	// 10. Extract JTI from the new Access Token and call DELETE /v1/auth/session to Logout
	jti, err := extractJTI(refreshResult.AccessToken)
	require.NoError(t, err)
	assert.NotEmpty(t, jti)

	logoutReq, err := http.NewRequest(http.MethodDelete, baseURL+"/v1/auth/session", nil)
	require.NoError(t, err)
	logoutReq.Header.Set("Authorization", "Bearer "+refreshResult.AccessToken)
	logoutReq.Header.Set("X-Refresh-Token", refreshResult.RefreshToken)

	logoutResp, err := authClient.Do(logoutReq)
	require.NoError(t, err)
	defer func() { _ = logoutResp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, logoutResp.StatusCode)

	// 11. Verify session deleted from database
	var postLogoutSessionCount int64
	err = db.Table("session_tokens").Where("refresh_token = ?", refreshResult.RefreshToken).Count(&postLogoutSessionCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), postLogoutSessionCount)

	// 12. Verify access token JTI is correctly placed on the Redis blacklist cache
	redisClient := goRedis.NewClient(&goRedis.Options{
		Addr: strings.TrimPrefix(redisURL, "redis://"),
	})
	defer func() { _ = redisClient.Close() }()

	blacklisted, err := redisClient.Exists(ctx, "blacklist:jti:"+jti).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), blacklisted)

	ttl, err := redisClient.TTL(ctx, "blacklist:jti:"+jti).Result()
	require.NoError(t, err)
	assert.True(t, ttl > 0)
	assert.True(t, ttl <= 15*time.Minute)

	// 13. Rejection of rotated refresh token (Double Refresh Attempt)
	doubleRefreshReq := dto.RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	}
	doubleRefreshReqBytes, err := json.Marshal(doubleRefreshReq)
	require.NoError(t, err)
	doubleRefreshResp, err := authClient.Post(baseURL+"/v1/auth/refresh", "application/json", bytes.NewBuffer(doubleRefreshReqBytes))
	require.NoError(t, err)
	defer func() { _ = doubleRefreshResp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, doubleRefreshResp.StatusCode)

	// 14. Rejection of expired refresh token
	// Insert an expired session into DB
	expiredToken := "expired-refresh-token-uuid"
	err = db.Exec("INSERT INTO session_tokens (id, admin_id, refresh_token, expires_at, is_persistent) VALUES (?, ?, ?, ?, ?)",
		domain.NewUUID(), adminUserID, expiredToken, time.Now().Add(-1*time.Hour), true).Error
	require.NoError(t, err)

	expiredRefreshReq := dto.RefreshRequest{
		RefreshToken: expiredToken,
	}
	expiredRefreshReqBytes, err := json.Marshal(expiredRefreshReq)
	require.NoError(t, err)
	expiredRefreshResp, err := authClient.Post(baseURL+"/v1/auth/refresh", "application/json", bytes.NewBuffer(expiredRefreshReqBytes))
	require.NoError(t, err)
	defer func() { _ = expiredRefreshResp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, expiredRefreshResp.StatusCode)

	// 15. Empty/Invalid refresh token
	emptyRefreshReq := dto.RefreshRequest{
		RefreshToken: "",
	}
	emptyRefreshReqBytes, err := json.Marshal(emptyRefreshReq)
	require.NoError(t, err)
	emptyRefreshResp, err := authClient.Post(baseURL+"/v1/auth/refresh", "application/json", bytes.NewBuffer(emptyRefreshReqBytes))
	require.NoError(t, err)
	defer func() { _ = emptyRefreshResp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, emptyRefreshResp.StatusCode)

	// 16. Malformed JSON Body
	malformedResp, err := authClient.Post(baseURL+"/v1/auth/refresh", "application/json", bytes.NewBuffer([]byte("{invalid-json}")))
	require.NoError(t, err)
	defer func() { _ = malformedResp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, malformedResp.StatusCode)

	// 17. Rejection of blacklisted access token JTI validation
	isJtiBlacklisted, err := redisClient.Exists(ctx, "blacklist:jti:"+jti).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), isJtiBlacklisted)
}
