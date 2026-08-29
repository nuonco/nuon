package nuonidentityprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

type Config struct{}

var componentType = component.MustNewType("nuonidentity")

func NewFactory() processor.Factory {
	return processor.NewFactory(
		componentType,
		createDefaultConfig,
		processor.WithLogs(createLogsProcessor, component.StabilityLevelDevelopment),
		processor.WithMetrics(createMetricsProcessor, component.StabilityLevelDevelopment),
		processor.WithTraces(createTracesProcessor, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createLogsProcessor(ctx context.Context, settings processor.Settings, cfg component.Config, next consumer.Logs) (processor.Logs, error) {
	return processorhelper.NewLogs(ctx, settings, cfg, next, func(ctx context.Context, logs plog.Logs) (plog.Logs, error) {
		if err := processLogs(ctx, logs); err != nil {
			return logs, err
		}
		return logs, nil
	})
}

func createMetricsProcessor(ctx context.Context, settings processor.Settings, cfg component.Config, next consumer.Metrics) (processor.Metrics, error) {
	return processorhelper.NewMetrics(ctx, settings, cfg, next, func(ctx context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
		if err := processMetrics(ctx, metrics); err != nil {
			return metrics, err
		}
		return metrics, nil
	})
}

func createTracesProcessor(ctx context.Context, settings processor.Settings, cfg component.Config, next consumer.Traces) (processor.Traces, error) {
	return processorhelper.NewTraces(ctx, settings, cfg, next, func(ctx context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
		if err := processTraces(ctx, traces); err != nil {
			return traces, err
		}
		return traces, nil
	})
}
