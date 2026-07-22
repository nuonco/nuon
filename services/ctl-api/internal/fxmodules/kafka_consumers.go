package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/consumer"
)

// KafkaConsumersModule wires every Kafka consumer. They run together in the
// dedicated `consumer` command (one process / deployment); add new consumers
// (logs, etc.) here. Each consumer no-ops unless KAFKA_ENABLED.
var KafkaConsumersModule = fx.Module("kafka-consumers",
	fx.Provide(consumer.NewHeartbeatConsumer),
	fx.Invoke(func(*consumer.HeartbeatConsumer) {}),
)
