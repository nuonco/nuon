package fxmodules

import (
	"go.uber.org/fx"

	runnersconsumer "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/consumer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
)

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
