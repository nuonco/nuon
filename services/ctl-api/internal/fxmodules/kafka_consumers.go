package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/consumer"
)

// KafkaConsumersModule wires every Kafka consumer. Which of them actually run is
// decided by `consumer --name`, not by what's registered here: deployed, each
// pod runs exactly one (its own deployment, resources, and consumer group);
// locally `--name=all` runs them together. Each also no-ops unless KAFKA_ENABLED.
//
// So registering a new consumer here is safe — it does not join any existing
// deployment. What it does need is its own entry in `consumer.Names()` plus a
// deployment instance in the ctl-api chart's `consumer.instances`, or nothing
// will ever select it.
var KafkaConsumersModule = fx.Module("kafka-consumers",
	fx.Provide(consumer.NewHeartbeatConsumer),
	fx.Provide(consumer.NewOtelLogsConsumer),
	fx.Provide(consumer.NewDLQConsumer),
	fx.Invoke(func(*consumer.HeartbeatConsumer) {}),
	fx.Invoke(func(*consumer.OtelLogsConsumer) {}),
	fx.Invoke(func(*consumer.DLQConsumer) {}),
)
