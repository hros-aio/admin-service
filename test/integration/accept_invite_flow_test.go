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
	"github.com/hros/admin-service/internal/application/usecase"
	"github.com/hros/admin-service/internal/config"
	"github.com/hros/admin-service/internal/domain"
	authInfra "github.com/hros/admin-service/internal/infrastructure/auth"
	authCache "github.com/hros/admin-service/internal/infrastructure/cache"
	authRepo "github.com/hros/admin-service/internal/infrastructure/repository/auth"
	"github.com/hros/admin-service/internal/platform/database"
	httpPlatform "github.com/hros/admin-service/internal/platform/http"
	platformRedis "github.com/hros/admin-service/internal/platform/redis"
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

func TestAcceptInviteFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Setup PostgreSQL container using testcontainers-go
	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("hros_admin_invite_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		err := postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// 2. Execute SQL migrations (000001 through 000004)
	migDir := findMigrationsDir(t)
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))
	runMigrationSQLFile(t, db, filepath.Join(migDir, "000004_create_invite_tokens.up.sql"))

	// 3. Seed required role
	roleID := domain.NewUUID()
	err = db.Exec("INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		roleID, "Standard Admin", "Standard Admin Role", false).Error
	require.NoError(t, err)

	// 4. Seed active inviter (Super Admin)
	inviterID := domain.NewUUID()
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		inviterID, "Super Admin", "super@hros.com", "dummyhash", roleID, "active").Error
	require.NoError(t, err)

	// 5. Seed pending invitee
	inviteeID := domain.NewUUID()
	inviteeEmail := "invitee@hros.com"
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		inviteeID, "Pending Admin", inviteeEmail, "pendinghash", roleID, "pending").Error
	require.NoError(t, err)

	// 6. Seed a valid, unconsumed invite token
	validToken := "valid-token-123"
	err = db.Exec("INSERT INTO invite_tokens (id, admin_id, token, inviter_id, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		domain.NewUUID(), inviteeID, validToken, inviterID, time.Now().Add(24*time.Hour), time.Now()).Error
	require.NoError(t, err)

	// 7. Seed an expired invite token associated with a different pending admin
	expiredToken := "expired-token-456"
	expiredInviteeID := domain.NewUUID()
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		expiredInviteeID, "Expired Admin", "expired@hros.com", "pendinghash", roleID, "pending").Error
	require.NoError(t, err)

	err = db.Exec("INSERT INTO invite_tokens (id, admin_id, token, inviter_id, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		domain.NewUUID(), expiredInviteeID, expiredToken, inviterID, time.Now().Add(-1*time.Hour), time.Now().Add(-49*time.Hour)).Error
	require.NoError(t, err)

	// 8. Setup Redis testcontainer (using existing session persistence flow helper)
	redisContainer, redisURL, err := runRedisContainer(ctx)
	require.NoError(t, err)
	defer func() {
		err := redisContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// 9. Setup Kafka SyncProducer Mock to expect exactly 1 message
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	// 10. Setup and start the Fx server
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
				AppName:       "admin-service-invite-test",
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
		fx.Provide(authRepo.NewGormInviteTokenRepository),
		fx.Provide(platformRedis.NewRedisClient),
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
		fx.Provide(func(p *kafkaProducer.NotificationKafkaProducer) usecase.NotificationPublisher { return p }),
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

	// 11. Run Expired Token Scenario (T024)
	t.Run("ExpiredToken", func(t *testing.T) {
		reqPayload := dto.AcceptInviteRequest{
			Token:                expiredToken,
			Password:             "SecurePassword123!",
			PasswordConfirmation: "SecurePassword123!",
		}
		bodyBytes, err := json.Marshal(reqPayload)
		require.NoError(t, err)

		resp, err := httpClient.Post(baseURL+"/v1/auth/accept-invite", "application/json", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp sharedErrors.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "INVITE_EXPIRED", errorResp.Code)
		_ = resp.Body.Close()

		// Verify database state: status remains pending, consumed_at remains null
		var dbUser struct {
			Status string `gorm:"column:status"`
		}
		err = db.Table("admin_users").Where("id = ?", expiredInviteeID).First(&dbUser).Error
		require.NoError(t, err)
		assert.Equal(t, "pending", dbUser.Status)

		var dbToken struct {
			ConsumedAt *time.Time `gorm:"column:consumed_at"`
		}
		err = db.Table("invite_tokens").Where("token = ?", expiredToken).First(&dbToken).Error
		require.NoError(t, err)
		assert.Nil(t, dbToken.ConsumedAt)
	})

	// 12. Run Happy Path Account Activation Flow (T023)
	t.Run("HappyPath", func(t *testing.T) {
		newPassword := "ActivatedPassword123!"
		reqPayload := dto.AcceptInviteRequest{
			Token:                validToken,
			Password:             newPassword,
			PasswordConfirmation: newPassword,
		}
		bodyBytes, err := json.Marshal(reqPayload)
		require.NoError(t, err)

		resp, err := httpClient.Post(baseURL+"/v1/auth/accept-invite", "application/json", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var successResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&successResp)
		require.NoError(t, err)
		assert.Equal(t, "Account activated successfully.", successResp["message"])
		_ = resp.Body.Close()

		// Verify database state: status is active, password_hash updated
		var dbUser struct {
			Status       string `gorm:"column:status"`
			PasswordHash string `gorm:"column:password_hash"`
		}
		err = db.Table("admin_users").Where("id = ?", inviteeID).First(&dbUser).Error
		require.NoError(t, err)
		assert.Equal(t, "active", dbUser.Status)
		assert.NotEmpty(t, dbUser.PasswordHash)

		err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(newPassword))
		assert.NoError(t, err, "password hash must match the new activated password")

		// Verify invite token is consumed
		var dbToken struct {
			ConsumedAt *time.Time `gorm:"column:consumed_at"`
		}
		err = db.Table("invite_tokens").Where("token = ?", validToken).First(&dbToken).Error
		require.NoError(t, err)
		assert.NotNil(t, dbToken.ConsumedAt, "invite token consumed_at must be populated")
	})
}
