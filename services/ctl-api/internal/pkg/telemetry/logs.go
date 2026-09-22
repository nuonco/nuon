package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
)

func NewLifecycleLogCore(lc fx.Lifecycle, cfg *Config) (zapcore.Core, error) {
	if cfg.Endpoint == "" {
		return zapcore.NewNopCore(), nil
	}
	endpoint, err := url.JoinPath(cfg.Endpoint, "v1/logs")
	if err != nil {
		return nil, fmt.Errorf("configure OTEL log endpoint: %w", err)
	}
	exporter, err := otlploghttp.New(context.Background(), otlploghttp.WithEndpointURL(endpoint))
	if err != nil {
		return nil, fmt.Errorf("unable to configure OTLP log exporter: %w", err)
	}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(cfg.Resource),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return provider.Shutdown(ctx)
		},
	})
	return &lifecycleLogCore{core: otelzap.NewCore("github.com/nuonco/nuon/ctl-api/lifecycle", otelzap.WithLoggerProvider(provider))}, nil
}

type lifecycleLogCore struct {
	core   zapcore.Core
	fields []zapcore.Field
}

func (c *lifecycleLogCore) Enabled(level zapcore.Level) bool {
	return level >= zapcore.InfoLevel && c.core.Enabled(level)
}

func (c *lifecycleLogCore) With(fields []zapcore.Field) zapcore.Core {
	selected := append([]zapcore.Field(nil), c.fields...)
	return &lifecycleLogCore{core: c.core, fields: append(selected, lifecycleFields(fields)...)}
}

func (c *lifecycleLogCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if entry.Message == "flow telemetry" && c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

func (c *lifecycleLogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range c.fields {
		field.AddTo(encoder)
	}
	for _, field := range lifecycleFields(fields) {
		field.AddTo(encoder)
	}
	event, _ := encoder.Fields["flow_event"].(string)
	switch event {
	case "workflow.created", "workflow.started", "workflow.completed", "workflow.failed", "workflow.cancelled",
		"step.started", "step.completed", "step.errored", "step.cancelled", "step.awaiting_retry", "step.awaiting_approval", "step.approval_resolved",
		"step_group.started", "step_group.completed", "step_group.failed", "step_group.cancelled",
		"step_generation.started", "step_generation.completed", "step_generation.failed", "step_generation.cancelled",
		"drift.detected", "app_config.synced", "install.config_updated", "install.config_update_failed",
		"component.unhealthy", "component.recovered", "install.degraded", "install.recovered",
		"runner_job.queued", "runner_job.available", "runner_job.in-progress", "runner_job.finished", "runner_job.failed", "runner_job.timed-out", "runner_job.cancelled", "runner_job.not-attempted", "runner_job.unknown":
	default:
		return nil
	}
	body, err := json.Marshal(encoder.Fields)
	if err != nil {
		return err
	}
	entry.Message = string(body)
	entry.Stack = ""
	return c.core.Write(entry, nil)
}

func (c *lifecycleLogCore) Sync() error {
	return c.core.Sync()
}

func lifecycleFields(fields []zapcore.Field) []zapcore.Field {
	selected := make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		switch field.Key {
		case "flow_event", "source", "org_id", "org_name", "app_id", "install_id", "install_name",
			"component_id", "component_name", "sandbox_id", "workflow_id", "workflow_type",
			"step_id", "step_name", "step_group_id", "queue_id", "queue_signal_id",
			"runner_job_id", "runner_id", "deploy_id", "build_id", "app_config_id",
			"owner_id", "owner_type", "operation", "stage", "signal_type", "phase", "status",
			"job_type", "job_operation", "health", "previous_health", "install_health", "install_previous_health":
			if field.Type != zapcore.StringType || len(field.String) > 512 {
				continue
			}
		case "step_idx", "group_idx", "group_retry_idx", "retry_index", "max_retries",
			"unhealthy_component_count", "degraded_component_count", "duration_ms", "signal_duration_ms":
			switch field.Type {
			case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
			default:
				continue
			}
		case "plan_only":
			if field.Type != zapcore.BoolType {
				continue
			}
		default:
			continue
		}
		selected = append(selected, field)
	}
	return selected
}
