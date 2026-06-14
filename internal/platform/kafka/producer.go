package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/hros/admin-service/internal/config"
	"go.uber.org/fx"
)

// NewKafkaProducer initializes the Kafka sync producer.
func NewKafkaProducer(cfg *config.Config, _ *slog.Logger, lc fx.Lifecycle) (sarama.SyncProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.KafkaBrokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return producer.Close()
		},
	})

	return producer, nil
}
