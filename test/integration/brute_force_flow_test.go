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
	adapterHttp "github.com/hros/admin-service/internal/adapter/http"
	"github.com/hros/admin-service/internal/adapter/http/auth/dto"
	kafkaProducer "github.com/hros/admin-service/internal/adapter/kafka/producer"
	"github.com/hros/admin-service/internal/application"
	"github.com/hros/admin-service/internal/application/interfaces"
	"github.com/hros/admin-service/internal/config"
	"github.com/hros/admin-service/internal/domain"
	"github.com/hros/admin-service/internal/domain/events"
	authInfra "github.com/hros/admin-service/internal/infrastructure/auth"
	authCache "github.com/hros/admin-service/internal/infrastructure/cache"
	authRepo "github.com/hros/admin-service/internal/infrastructure/repository/auth"
	"github.com/hros/admin-service/internal/platform/database"
	httpPlatform "github.com/hros/admin-service/internal/platform/http"
	"github.com/hros/admin-service/internal/platform/redis"
	sharedErrors "github.com/hros/admin-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBruteForceFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Setup testcontainers PostgreSQL instance
	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("hros_admin"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
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
	adminEmail := "brute-admin@hros.com"
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		adminUserID, "Brute Test Admin", adminEmail, string(hashedPassword), roleID, "active").Error
	require.NoError(t, err)

	// 5. Setup Mock Kafka SyncProducer to capture the event
	var capturedMsg []byte
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(msg []byte) error {
		capturedMsg = msg
		return nil
	})

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
				AppName:       "admin-service-brute-force-test",
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
		_ = mockProducer.Close()
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

	// 7. Verify that successful login prior to 5 fails clears the counter.
	// First, simulate 3 failed login attempts.
	for i := 1; i <= 3; i++ {
		func() {
			badLoginReq := dto.LoginRequest{
				Email:    adminEmail,
				Password: "wrong-password",
			}
			badLoginReqBytes, _ := json.Marshal(badLoginReq)
			badResp, err := authClient.Post(baseURL+"/v1/auth/login", "application/json", bytes.NewBuffer(badLoginReqBytes))
			require.NoError(t, err)
			defer func() { _ = badResp.Body.Close() }()

			assert.Equal(t, http.StatusUnauthorized, badResp.StatusCode)
			var errResp sharedErrors.ErrorResponse
			err = json.NewDecoder(badResp.Body).Decode(&errResp)
			require.NoError(t, err)
			assert.Equal(t, "unauthorized", errResp.Code)
		}()
	}

	// Next, perform a successful login with valid credentials.
	loginReq := dto.LoginRequest{
		Email:    adminEmail,
		Password: "password123",
	}
	loginReqBytes, _ := json.Marshal(loginReq)
	resp, err := authClient.Post(baseURL+"/v1/auth/login", "application/json", bytes.NewBuffer(loginReqBytes))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp dto.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)

	// Since we logged in successfully, the counter must have been cleared to 0.
	// This means we now need a fresh 5 consecutive failed attempts to lock out the account.

	// 8. Simulate 5 failed logins to lock out the account.
	// 1st to 4th failed attempts should return HTTP 401 "unauthorized".
	for i := 1; i <= 4; i++ {
		func() {
			badLoginReq := dto.LoginRequest{
				Email:    adminEmail,
				Password: "wrong-password",
			}
			badLoginReqBytes, _ := json.Marshal(badLoginReq)
			badResp, err := authClient.Post(baseURL+"/v1/auth/login", "application/json", bytes.NewBuffer(badLoginReqBytes))
			require.NoError(t, err)
			defer func() { _ = badResp.Body.Close() }()

			assert.Equal(t, http.StatusUnauthorized, badResp.StatusCode)
			var errResp sharedErrors.ErrorResponse
			err = json.NewDecoder(badResp.Body).Decode(&errResp)
			require.NoError(t, err)
			assert.Equal(t, "unauthorized", errResp.Code)
		}()
	}

	// The 5th failed attempt should also return 401 "unauthorized",
	// but it must trigger the lockout and publish a Kafka event.
	badLoginReq5 := dto.LoginRequest{
		Email:    adminEmail,
		Password: "wrong-password",
	}
	badLoginReqBytes5, _ := json.Marshal(badLoginReq5)
	badResp5, err := authClient.Post(baseURL+"/v1/auth/login", "application/json", bytes.NewBuffer(badLoginReqBytes5))
	require.NoError(t, err)
	defer func() { _ = badResp5.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, badResp5.StatusCode)
	var errResp5 sharedErrors.ErrorResponse
	err = json.NewDecoder(badResp5.Body).Decode(&errResp5)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", errResp5.Code)

	// 9. Verify the 5th attempt triggered the Kafka email.send message
	assert.NotEmpty(t, capturedMsg, "Kafka message should have been published")
	var envelope kafkaProducer.EventEnvelope[events.EmailSendEvent]
	err = json.Unmarshal(capturedMsg, &envelope)
	require.NoError(t, err)

	assert.Equal(t, "email.send", envelope.Type)
	assert.Equal(t, "admin-service", envelope.Source)
	assert.Equal(t, 1, envelope.Version)
	assert.Equal(t, adminEmail, envelope.Data.To)
	assert.Equal(t, "Your account has been temporarily locked", envelope.Data.Subject)
	assert.Equal(t, "account_locked_notification", envelope.Data.Template)
	assert.Equal(t, adminEmail, envelope.Data.TemplateData["email"])
	assert.NotEmpty(t, envelope.Data.TemplateData["unlock_at"])

	// 10. Verify that the 6th login attempt returns 401 ACCOUNT_LOCKED (even with correct password)
	lockedLoginReq := dto.LoginRequest{
		Email:    adminEmail,
		Password: "password123", // Correct password
	}
	lockedLoginReqBytes, _ := json.Marshal(lockedLoginReq)
	lockedResp, err := authClient.Post(baseURL+"/v1/auth/login", "application/json", bytes.NewBuffer(lockedLoginReqBytes))
	require.NoError(t, err)
	defer func() { _ = lockedResp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, lockedResp.StatusCode)
	var lockedErrResp sharedErrors.ErrorResponse
	err = json.NewDecoder(lockedResp.Body).Decode(&lockedErrResp)
	require.NoError(t, err)
	assert.Equal(t, "ACCOUNT_LOCKED", lockedErrResp.Code)
	assert.Equal(t, "Account is temporarily locked", lockedErrResp.Message)
}
