package fxmodules

import (
	"go.uber.org/fx"

	runnersconsumer "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/consumer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
)

// KafkaConsumersModule wires every Kafka consumer, domain ones alongside the DLQ
// consumer from the runtime package. Which of them actually run is decided by
// `consumer --name`, not by what's registered here: each consumer keeps its own
// topic, consumer group and client, and a deployment runs one or more of them
// (the `otel` deployment runs otel-logs and otel-traces together). Locally
// `--name=all` runs them all in one process. Each also no-ops unless
// KAFKA_ENABLED.
//
// So registering a new consumer here is safe — it does not join any existing
// deployment. What it does need is its own entry in `consumer.Names()`, and
// either a new deployment instance in the ctl-api chart's `consumer.instances`
// or an added name on an existing instance's `--name`, or nothing will ever
// select it.
var KafkaConsumersModule = fx.Module("kafka-consumers",
	fx.Provide(runnersconsumer.NewHeartbeatConsumer),
	fx.Provide(runnersconsumer.NewOtelLogsConsumer),
	fx.Provide(runnersconsumer.NewOtelTracesConsumer),
	fx.Provide(consumer.NewDLQConsumer),
	fx.Invoke(func(*runnersconsumer.HeartbeatConsumer) {}),
	fx.Invoke(func(*runnersconsumer.OtelLogsConsumer) {}),
	fx.Invoke(func(*runnersconsumer.OtelTracesConsumer) {}),
	fx.Invoke(func(*consumer.DLQConsumer) {}),
)
