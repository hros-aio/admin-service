// Package kafka provides tools for kafka production and consumption.
package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/hros/admin-service/internal/config"
	"go.uber.org/fx"
)

// NewKafkaConsumerGroup initializes the Kafka consumer group.
func NewKafkaConsumerGroup(cfg *config.Config, logger *slog.Logger, lc fx.Lifecycle) (sarama.ConsumerGroup, error) {
	if !cfg.KafkaConsumeEnable {
		logger.Info("Kafka consumer is disabled (KAFKA_CONSUME_ENABLE is not true), skipping consumer group creation")
		return nil, nil
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Return.Errors = true

	groupID := fmt.Sprintf("%s-group", cfg.AppName)
	consumerGroup, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, groupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer group: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return consumerGroup.Close()
		},
	})

	return consumerGroup, nil
}
