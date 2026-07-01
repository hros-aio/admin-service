// Package http provides HTTP handlers and server setup.
package http

import (
	"context"
	"net/http"

	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// StatusUP represents a healthy dependency status.
	StatusUP = "UP"
	// StatusDOWN represents an unhealthy dependency status.
	StatusDOWN = "DOWN"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
	kafka sarama.SyncProducer
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *gorm.DB, redis *redis.Client, kafka sarama.SyncProducer) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
		kafka: kafka,
	}
}

// HealthStatus represents the health of the application and its dependencies.
type HealthStatus struct {
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

// Check returns the health status of the application and its dependencies.
func (h *HealthHandler) Check(c echo.Context) error {
	status := "OK"
	deps := make(map[string]string)

	// Check Postgres
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		deps["postgres"] = StatusDOWN
		status = StatusDOWN
	} else {
		deps["postgres"] = StatusUP
	}

	// Check Redis
	if err := h.redis.Ping(context.Background()).Err(); err != nil {
		deps["redis"] = StatusDOWN
		status = StatusDOWN
	} else {
		deps["redis"] = StatusUP
	}

	// Check Kafka
	if h.kafka == nil {
		deps["kafka"] = "DISABLED"
	} else {
		deps["kafka"] = StatusUP
	}

	resp := HealthStatus{
		Status:       status,
		Version:      "0.1.0",
		Dependencies: deps,
	}

	if status == StatusDOWN {
		return c.JSON(http.StatusServiceUnavailable, resp)
	}

	return c.JSON(http.StatusOK, resp)
}
