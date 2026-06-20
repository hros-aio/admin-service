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
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/alicebob/miniredis/v2"
	adapterHttp "github.com/hros/admin-service/internal/adapter/http"
	"github.com/hros/admin-service/internal/adapter/http/auth/dto"
	"github.com/hros/admin-service/internal/application"
	"github.com/hros/admin-service/internal/config"
	"github.com/hros/admin-service/internal/domain"
	authInfra "github.com/hros/admin-service/internal/infrastructure/auth"
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

// Mock Sarama structures for integration testing DI resolution

type mockConsumerGroup struct{}

func (m *mockConsumerGroup) Consume(_ context.Context, _ []string, _ sarama.ConsumerGroupHandler) error {
	return nil
}
func (m *mockConsumerGroup) Errors() <-chan error {
	return make(chan error)
}
func (m *mockConsumerGroup) Close() error {
	return nil
}
func (m *mockConsumerGroup) Pause(_ map[string][]int32)  {}
func (m *mockConsumerGroup) Resume(_ map[string][]int32) {}
func (m *mockConsumerGroup) PauseAll()                   {}
func (m *mockConsumerGroup) ResumeAll()                  {}

// findMigrationsDir walks up the directory tree to locate the migrations folder.
func findMigrationsDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		path := filepath.Join(dir, "migrations")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not find migrations directory")
	return ""
}

// runSQLFile parses and executes SQL migration scripts statement-by-statement.
func runSQLFile(t *testing.T, db *gorm.DB, filepath string) {
	t.Helper()
	content, err := os.ReadFile(filepath)
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	statements := strings.Split(strings.Join(cleanLines, "\n"), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		err := db.Exec(stmt).Error
		require.NoError(t, err, "failed to execute statement: %s", stmt)
	}
}

// getFreePort allocates a free TCP port dynamically from the OS.
func getFreePort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func TestAuthFlow(t *testing.T) {
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

	// 2. Execute database migrations
	migDir := findMigrationsDir(t)
	runSQLFile(t, db, filepath.Join(migDir, "000001_init.up.sql"))
	runSQLFile(t, db, filepath.Join(migDir, "000002_create_auth_tables.up.sql"))

	// 3. Seed active Super Admin user
	roleID := domain.NewUUID()
	err = db.Exec("INSERT INTO roles (id, name, description, is_system_role) VALUES (?, ?, ?, ?)",
		roleID, "Super Admin", "Super Admin Role", true).Error
	require.NoError(t, err)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	require.NoError(t, err)

	adminUserID := domain.NewUUID()
	err = db.Exec("INSERT INTO admin_users (id, name, email, password_hash, role_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		adminUserID, "Test Admin", "test-admin@hros.com", string(hashedPassword), roleID, "active").Error
	require.NoError(t, err)

	// 4. Start in-memory miniredis instance
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

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
				AppName:       "admin-service-integration-test",
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
		fx.Provide(func() (sarama.SyncProducer, error) {
			return mocks.NewSyncProducer(t, nil), nil
		}),
		fx.Provide(func() (sarama.ConsumerGroup, error) {
			return &mockConsumerGroup{}, nil
		}),
		authInfra.Module,
		application.Module,
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

	// 6. Execute POST /v1/auth/login with valid credentials
	loginReq := dto.LoginRequest{
		Email:    "test-admin@hros.com",
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
	assert.NotEmpty(t, loginResp.RefreshToken)

	// 7. Execute DELETE /v1/auth/session with valid token
	logoutReq, _ := http.NewRequest(http.MethodDelete, baseURL+"/v1/auth/session", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+loginResp.RefreshToken)
	logoutResp, err := authClient.Do(logoutReq)
	require.NoError(t, err)
	defer func() { _ = logoutResp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, logoutResp.StatusCode)

	// 8. Execute POST /v1/auth/login with invalid credentials (401 error path)
	badLoginReq := dto.LoginRequest{
		Email:    "test-admin@hros.com",
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
}
