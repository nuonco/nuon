package slog

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	runnerlog "github.com/nuonco/nuon/pkg/runner/log"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

const (
	defaultOTLPLogsEndpointTmpl string = "%s/v1/log-streams/%s/logs"

	auditCollectorEndpoint string = "http://127.0.0.1:4318/v1/logs"
)

type auditProcessor struct {
	next log.Processor
}

func (p *auditProcessor) Enabled(ctx context.Context, params log.EnabledParameters) bool {
	return p.next.Enabled(ctx, params)
}

func (p *auditProcessor) OnEmit(ctx context.Context, record *log.Record) error {
	isAudit := false
	record.WalkAttributes(func(attr otellog.KeyValue) bool {
		isAudit = attr.Key == runnerlog.AuditAttr &&
			attr.Value.Kind() == otellog.KindString &&
			attr.Value.AsString() == runnerlog.AuditAttrValue
		return !isAudit
	})
	if !isAudit {
		return nil
	}

	return p.next.OnEmit(ctx, record)
}

func (p *auditProcessor) Shutdown(ctx context.Context) error {
	return p.next.Shutdown(ctx)
}

func (p *auditProcessor) ForceFlush(ctx context.Context) error {
	return p.next.ForceFlush(ctx)
}

func NewOTELProvider(cfg *runnerconfig.Config, set *settings.Settings, logStreamID string) (*log.LoggerProvider, error) {
	opts := []log.LoggerProviderOption{
		log.WithResource(getResource(set, logStreamID)),
	}

	if set.EnableLogging {
		processor, err := newAPIProcessor(cfg, logStreamID)
		if err != nil {
			return nil, err
		}
		opts = append(opts, log.WithProcessor(processor))
	}

	if cfg.OTELAuditExportEnabled {
		processor, err := newAuditCollectorProcessor()
		if err != nil {
			return nil, err
		}
		opts = append(opts, log.WithProcessor(processor))
	}

	return log.NewLoggerProvider(opts...), nil
}

func newAPIProcessor(cfg *runnerconfig.Config, logStreamID string) (log.Processor, error) {
	url := fmt.Sprintf(defaultOTLPLogsEndpointTmpl, cfg.RunnerAPIURL, logStreamID)
	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpointURL(url),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + cfg.RunnerAPIToken,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize otlp log exporter: %w", err)
	}

	return newBatchProcessor(exporter), nil
}

func newAuditCollectorProcessor() (log.Processor, error) {
	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpointURL(auditCollectorEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize audit collector log exporter: %w", err)
	}

	return &auditProcessor{next: newBatchProcessor(exporter)}, nil
}

func newBatchProcessor(exporter log.Exporter) log.Processor {
	return log.NewBatchProcessor(exporter,
		log.WithExportMaxBatchSize(25),
		log.WithExportInterval(time.Second*5))
}
