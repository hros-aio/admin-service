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
	platformRedis "github.com/hros/admin-service/internal/platform/redis"
	sharedErrors "github.com/hros/admin-service/internal/shared/errors"
	redis "github.com/redis/go-redis/v9"
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

func TestPasswordResetFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Setup testcontainers PostgreSQL instance
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("hros_admin_reset"),
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

	// 2. Execute database migrations
	migDir := findMigrationsDir(t)
	runSQLFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runSQLFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))

	// 3. Seed active Standard Admin user
	roleID := domain.NewUUID()
	err = db.Exec("INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		roleID, "Standard Admin", "Standard Admin Role", false).Error
	require.NoError(t, err)

	oldPassword := "OldSecurePassword1!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(oldPassword), 12)
	require.NoError(t, err)

	adminUserID := domain.NewUUID()
	adminEmail := "integration-reset@hros.com"
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		adminUserID, "Reset User", adminEmail, string(hashedPassword), roleID, "active").Error
	require.NoError(t, err)

	// 4. Seed active session tokens for the admin user
	err = db.Exec("INSERT INTO session_tokens (id, admin_id, refresh_token, expires_at, is_persistent) VALUES (?, ?, ?, ?, ?)",
		domain.NewUUID(), adminUserID, "session-token-1", time.Now().Add(1*time.Hour), true).Error
	require.NoError(t, err)

	err = db.Exec("INSERT INTO session_tokens (id, admin_id, refresh_token, expires_at, is_persistent) VALUES (?, ?, ?, ?, ?)",
		domain.NewUUID(), adminUserID, "session-token-2", time.Now().Add(2*time.Hour), true).Error
	require.NoError(t, err)

	// Verify session count is initially 2
	var initialSessionCount int64
	err = db.Table("session_tokens").Where("admin_id = ?", adminUserID).Count(&initialSessionCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(2), initialSessionCount)

	// 5. Setup testcontainers Redis instance
	redisContainer, redisURL, err := runRedisContainer(ctx)
	require.NoError(t, err)
	defer func() {
		err := redisContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	// 6. Bootstrap the Fx Echo application with custom overridden config
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
				AppName:       "admin-service-reset-test",
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
		fx.Provide(platformRedis.NewRedisClient),
		fx.Provide(authCache.NewRedisTokenBlacklist),
		fx.Provide(authCache.NewRedisBruteForceCache),
		fx.Provide(authCache.NewRedisMFACache),
		fx.Provide(authCache.NewRedisPasswordResetCache),
		fx.Provide(func() (sarama.SyncProducer, error) {
			return mockProducer, nil
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
	httpClient := &http.Client{Timeout: 15 * time.Second}

	// 7. Request password reset via HTTP API
	reqPayload := dto.PasswordResetRequest{
		Email: adminEmail,
	}
	payloadBytes, err := json.Marshal(reqPayload)
	require.NoError(t, err)

	resp, err := httpClient.Post(baseURL+"/v1/auth/password-reset/request", "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Read and verify success message
	var successResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&successResp)
	require.NoError(t, err)
	assert.Equal(t, "If an account exists for that email, a reset link has been sent.", successResp["message"])
	_ = resp.Body.Close()

	// 8. Capture generated token from Redis testcontainer
	redisOpts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)
	rClient := redis.NewClient(redisOpts)
	defer func() { _ = rClient.Close() }()

	var keys []string
	iter := rClient.Scan(ctx, 0, "auth:reset_token:*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	require.NoError(t, iter.Err())
	require.Len(t, keys, 1, "exactly one reset token should be cached in Redis")

	tokenKey := keys[0]
	cachedAdminID, err := rClient.Get(ctx, tokenKey).Result()
	require.NoError(t, err)
	assert.Equal(t, adminUserID, cachedAdminID)

	token := strings.TrimPrefix(tokenKey, "auth:reset_token:")

	// 9. Verify confirming with a weak password fails with HTTP 422
	weakConfirmPayload := dto.PasswordResetConfirmRequest{
		Token:                token,
		Password:             "weak",
		PasswordConfirmation: "weak",
	}
	weakBytes, err := json.Marshal(weakConfirmPayload)
	require.NoError(t, err)

	weakResp, err := httpClient.Post(baseURL+"/v1/auth/password-reset/confirm", "application/json", bytes.NewBuffer(weakBytes))
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, weakResp.StatusCode)

	var weakErrorResp sharedErrors.ErrorResponse
	err = json.NewDecoder(weakResp.Body).Decode(&weakErrorResp)
	require.NoError(t, err)
	assert.Equal(t, "PASSWORD_WEAK", weakErrorResp.Code)
	_ = weakResp.Body.Close()

	// 10. Execute the confirmation with a valid password
	newPassword := "NewSecurePassword2!"
	confirmPayload := dto.PasswordResetConfirmRequest{
		Token:                token,
		Password:             newPassword,
		PasswordConfirmation: newPassword,
	}
	confirmBytes, err := json.Marshal(confirmPayload)
	require.NoError(t, err)

	confirmResp, err := httpClient.Post(baseURL+"/v1/auth/password-reset/confirm", "application/json", bytes.NewBuffer(confirmBytes))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, confirmResp.StatusCode)

	var confirmSuccessResp map[string]string
	err = json.NewDecoder(confirmResp.Body).Decode(&confirmSuccessResp)
	require.NoError(t, err)
	assert.Equal(t, "Password updated successfully.", confirmSuccessResp["message"])
	_ = confirmResp.Body.Close()

	// 11. Verify that database hash updates
	var updatedUser struct {
		PasswordHash string `gorm:"column:password_hash"`
	}
	err = db.Table("admin_users").Where("id = ?", adminUserID).First(&updatedUser).Error
	require.NoError(t, err)
	assert.NotEqual(t, string(hashedPassword), updatedUser.PasswordHash)

	// Verify new password hash works
	err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte(newPassword))
	assert.NoError(t, err)

	// 12. Verify that existing active sessions are deleted
	var postResetSessionCount int64
	err = db.Table("session_tokens").Where("admin_id = ?", adminUserID).Count(&postResetSessionCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), postResetSessionCount, "all active sessions must be cleared")

	// 13. Verify that already consumed token returns a 400 TOKEN_USED error
	reuseResp, err := httpClient.Post(baseURL+"/v1/auth/password-reset/confirm", "application/json", bytes.NewBuffer(confirmBytes))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, reuseResp.StatusCode)

	var reuseErrorResp sharedErrors.ErrorResponse
	err = json.NewDecoder(reuseResp.Body).Decode(&reuseErrorResp)
	require.NoError(t, err)
	assert.Equal(t, "TOKEN_USED", reuseErrorResp.Code)
	_ = reuseResp.Body.Close()

	// 14. Verify that an expired/nonexistent token returns a 400 TOKEN_EXPIRED error
	expiredConfirmPayload := dto.PasswordResetConfirmRequest{
		Token:                "nonexistent-expired-token",
		Password:             "AnotherSecurePass3!",
		PasswordConfirmation: "AnotherSecurePass3!",
	}
	expiredBytes, err := json.Marshal(expiredConfirmPayload)
	require.NoError(t, err)

	expiredResp, err := httpClient.Post(baseURL+"/v1/auth/password-reset/confirm", "application/json", bytes.NewBuffer(expiredBytes))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, expiredResp.StatusCode)

	var expiredErrorResp sharedErrors.ErrorResponse
	err = json.NewDecoder(expiredResp.Body).Decode(&expiredErrorResp)
	require.NoError(t, err)
	assert.Equal(t, "TOKEN_EXPIRED", expiredErrorResp.Code)
	_ = expiredResp.Body.Close()
}
