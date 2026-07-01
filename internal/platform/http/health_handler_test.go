package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/IBM/sarama/mocks"
	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHealthHandler_Check(t *testing.T) {
	// Setup Postgres Mock
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Expect initial ping from gorm.Open
	mock.ExpectPing()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	// Setup Redis Mock
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Setup Kafka Mock
	kafkaProducer := mocks.NewSyncProducer(t, nil)

	handler := NewHealthHandler(gormDB, redisClient, kafkaProducer)

	t.Run("success", func(t *testing.T) {
		mock.ExpectPing()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Check(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"status":"OK"`)
	})

	t.Run("db_down", func(t *testing.T) {
		// Close DB to simulate failure
		_ = db.Close()

		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Check(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Contains(t, rec.Body.String(), `"postgres":"DOWN"`)
	})
}

func TestHealthHandler_Check_KafkaDisabled(t *testing.T) {
	// Setup Postgres Mock
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectPing()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	// Setup Redis Mock
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Pass nil for Kafka SyncProducer to simulate disabled state
	handler := NewHealthHandler(gormDB, redisClient, nil)

	mock.ExpectPing()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.Check(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"OK"`)
	assert.Contains(t, rec.Body.String(), `"kafka":"DISABLED"`)
}
