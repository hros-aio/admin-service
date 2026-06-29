package integration

import (
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
	"net/url"
	"os"
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
	authInfra "github.com/hros/admin-service/internal/infrastructure/auth"
	authCache "github.com/hros/admin-service/internal/infrastructure/cache"
	authRepo "github.com/hros/admin-service/internal/infrastructure/repository/auth"
	"github.com/hros/admin-service/internal/platform/database"
	httpPlatform "github.com/hros/admin-service/internal/platform/http"
	"github.com/hros/admin-service/internal/platform/redis"
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

type testSSOClient struct {
	email      string
	identityID string
}

func (c *testSSOClient) ExchangeCode(ctx context.Context, provider string, code string) (*interfaces.SSOUserProfile, error) {
	return &interfaces.SSOUserProfile{
		Email:      c.email,
		IdentityID: c.identityID,
		Provider:   provider,
	}, nil
}

func runMigrationFile(t *testing.T, db *gorm.DB, filepath string) {
	t.Helper()
	content, err := os.ReadFile(filepath)
	require.NoError(t, err)

	err = db.Exec(string(content)).Error
	require.NoError(t, err, "failed to execute migration: %s", filepath)
}

func TestSSOFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Setup testcontainers PostgreSQL instance
	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("hros_admin"),
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

	// Connect directly with GORM to run migrations and seed test data
	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// 2. Execute database migrations
	migDir := findMigrationsDir(t)
	runMigrationFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runMigrationFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))
	runMigrationFile(t, db, filepath.Join(migDir, "000003_add_mfa_to_admin_users.up.sql"))
	runMigrationFile(t, db, filepath.Join(migDir, "000004_create_invite_tokens.up.sql"))
	runMigrationFile(t, db, filepath.Join(migDir, "000005_add_sso_to_admin_users.up.sql"))

	// 3. Seed active Standard Admin role
	roleID := domain.NewUUID()
	err = db.Exec("INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		roleID, "Standard Admin", "Standard Admin Role", false).Error
	require.NoError(t, err)

	// Seed Admin user mapping sso-admin@hros.com to sso-identity-123
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	require.NoError(t, err)

	adminUserID := domain.NewUUID()
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status, sso_identity_id, sso_provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		adminUserID, "SSO Admin User", "sso-admin@hros.com", string(hashedPassword), roleID, "active", "sso-identity-123", "google").Error
	require.NoError(t, err)

	// 4. Setup Redis container
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		err := redisContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	redisHost, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)
	redisURL := fmt.Sprintf("redis://%s:%s", redisHost, redisPort.Port())

	// 5. Setup test SSO Client mock
	testClient := &testSSOClient{
		email:      "sso-admin@hros.com",
		identityID: "sso-identity-123",
	}

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
				AppName:              "admin-service-sso-integration-test",
				Env:                  "test",
				Port:                 testPort,
				DBURL:                connStr,
				RedisURL:             redisURL,
				KafkaBrokers:         []string{"localhost:9092"},
				LogLevel:             "debug",
				JWTPrivateKey:        string(pemBytes),
				SSOGoogleClientID:    "google-client-id",
				SSOGoogleRedirectURL: "https://hros.io/callback",
				SSOGoogleAuthURL:     "https://accounts.google.com/o/oauth2/auth",
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
		fx.Decorate(func(defaultClient interfaces.SSOClient) interfaces.SSOClient {
			return testClient
		}),
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
	authClient := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Do not follow redirect so we can extract redirect location header
			return http.ErrUseLastResponse
		},
	}

	t.Run("Initiate SSO and verify redirect + Redis state creation", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, baseURL+"/v1/auth/sso/initiate?provider=google", nil)
		require.NoError(t, err)

		resp, err := authClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		loc := resp.Header.Get("Location")
		require.NotEmpty(t, loc)

		u, err := url.Parse(loc)
		require.NoError(t, err)
		assert.Equal(t, "accounts.google.com", u.Host)
		assert.Equal(t, "/o/oauth2/auth", u.Path)

		q := u.Query()
		assert.Equal(t, "google-client-id", q.Get("client_id"))
		assert.Equal(t, "https://hros.io/callback", q.Get("redirect_uri"))
		assert.Equal(t, "code", q.Get("response_type"))
		assert.NotEmpty(t, q.Get("state"))
		assert.NotEmpty(t, q.Get("nonce"))
	})

	t.Run("Execute Callback - Success flow with registered user", func(t *testing.T) {
		// 1. Get a new state parameter by calling initiate
		reqInit, err := http.NewRequest(http.MethodGet, baseURL+"/v1/auth/sso/initiate?provider=google", nil)
		require.NoError(t, err)

		respInit, err := authClient.Do(reqInit)
		require.NoError(t, err)
		defer respInit.Body.Close()

		loc := respInit.Header.Get("Location")
		u, err := url.Parse(loc)
		require.NoError(t, err)
		state := u.Query().Get("state")
		require.NotEmpty(t, state)

		// Ensure mock provider has the correct registered email and identity ID
		testClient.email = "sso-admin@hros.com"
		testClient.identityID = "sso-identity-123"

		// 2. Call Callback handler using the generated state
		callbackURL := fmt.Sprintf("%s/v1/auth/sso/callback?code=mock-auth-code&state=%s&provider=google", baseURL, state)
		reqCallback, err := http.NewRequest(http.MethodGet, callbackURL, nil)
		require.NoError(t, err)
		reqCallback.Header.Set("Accept", "application/json")

		respCallback, err := authClient.Do(reqCallback)
		require.NoError(t, err)
		defer respCallback.Body.Close()

		assert.Equal(t, http.StatusOK, respCallback.StatusCode)

		var loginResp dto.LoginResponse
		err = json.NewDecoder(respCallback.Body).Decode(&loginResp)
		require.NoError(t, err)

		assert.NotEmpty(t, loginResp.AccessToken)
		assert.Empty(t, loginResp.RefreshToken) // Kept exclusively in HTTP-only cookie

		// Verify HTTP-only cookie is set
		cookies := respCallback.Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "refresh_token", cookies[0].Name)
		assert.NotEmpty(t, cookies[0].Value)
		assert.True(t, cookies[0].HttpOnly)

		// Verify session token was created in the database
		var sessionCount int64
		err = db.Table("session_tokens").Where("admin_id = ?", adminUserID).Count(&sessionCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), sessionCount)
	})

	t.Run("Execute Callback - Failure flow with unregistered user (401)", func(t *testing.T) {
		// 1. Get a new state parameter by calling initiate
		reqInit, err := http.NewRequest(http.MethodGet, baseURL+"/v1/auth/sso/initiate?provider=google", nil)
		require.NoError(t, err)

		respInit, err := authClient.Do(reqInit)
		require.NoError(t, err)
		defer respInit.Body.Close()

		loc := respInit.Header.Get("Location")
		u, err := url.Parse(loc)
		require.NoError(t, err)
		state := u.Query().Get("state")
		require.NotEmpty(t, state)

		// Configure mock provider to return an unregistered email address and unregistered identity ID
		testClient.email = "unregistered-user@hros.com"
		testClient.identityID = "unregistered-sso-id-999"

		// 2. Call Callback handler using the generated state
		callbackURL := fmt.Sprintf("%s/v1/auth/sso/callback?code=mock-auth-code&state=%s&provider=google", baseURL, state)
		reqCallback, err := http.NewRequest(http.MethodGet, callbackURL, nil)
		require.NoError(t, err)
		reqCallback.Header.Set("Accept", "application/json")

		respCallback, err := authClient.Do(reqCallback)
		require.NoError(t, err)
		defer respCallback.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, respCallback.StatusCode)

		var errResp map[string]interface{}
		err = json.NewDecoder(respCallback.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Equal(t, "NO_ACCOUNT_LINKED", errResp["code"])
	})
}
