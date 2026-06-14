package kafka

import (
	"log/slog"
	"os"
	"testing"

	"github.com/hros/admin-service/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestNewKafkaConsumerGroup(t *testing.T) {
	cfg := &config.Config{
		AppName:      "test-app",
		KafkaBrokers: []string{"localhost:9092"},
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("wiring", func(t *testing.T) {
		err := fx.ValidateApp(
			fx.Provide(func() *config.Config { return cfg }),
			fx.Provide(func() *slog.Logger { return log }),
			fx.Provide(NewKafkaConsumerGroup),
		)
		assert.NoError(t, err)
	})
}
