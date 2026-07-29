package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/consumer"
)

// KafkaConsumersModule wires every Kafka consumer. They run together in the
// dedicated `consumer` command, deployed as ctl-api-clickhouse-sink; add new
// sinks (logs, etc.) here. Each consumer no-ops unless KAFKA_ENABLED.
//
// Everything registered here writes to ClickHouse, which is what that deployment
// is named for. A consumer that does something else — reacting to events, audit
// trails — wants its own consumer group so it can lag and replay independently,
// so give it its own deployment and a flag to select which consumers run,
// mirroring `worker --namespace`. Registering it here instead would silently put
// it in the sink deployment and share the sink's group.
var KafkaConsumersModule = fx.Module("kafka-consumers",
	fx.Provide(consumer.NewHeartbeatConsumer),
	fx.Invoke(func(*consumer.HeartbeatConsumer) {}),
)
