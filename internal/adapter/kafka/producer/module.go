// Package producer provides Kafka event producer adapters for the admin service.
package producer

import "go.uber.org/fx"

// Module is the Uber Fx module for the Kafka producer adapter layer.
// It provides EmailKafkaProducer and NotificationKafkaProducer to the dependency graph.
var Module = fx.Module(
	"kafka-producer",
	fx.Provide(
		NewEmailKafkaProducer,
		NewNotificationKafkaProducer,
	),
)
