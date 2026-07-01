package kafka

import (
	"log/slog"
	"os"
	"testing"

	"github.com/hros/admin-service/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestNewKafkaProducer(t *testing.T) {
	cfg := &config.Config{
		AppName:      "test-app",
		KafkaBrokers: []string{"localhost:9092"},
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("wiring", func(t *testing.T) {
		err := fx.ValidateApp(
			fx.Provide(func() *config.Config { return cfg }),
			fx.Provide(func() *slog.Logger { return log }),
			fx.Provide(NewKafkaProducer),
		)
		assert.NoError(t, err)
	})
}

func TestNewKafkaProducer_Direct(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("disabled", func(t *testing.T) {
		cfg := &config.Config{
			AppName:            "test-app",
			KafkaBrokers:       []string{"localhost:9092"},
			KafkaProduceEnable: false,
		}
		// Passing nil for fx.Lifecycle since it shouldn't be touched when disabled.
		p, err := NewKafkaProducer(cfg, log, nil)
		assert.NoError(t, err)
		assert.Nil(t, p)
	})
}
