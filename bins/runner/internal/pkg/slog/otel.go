package slog

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

const (
	defaultOTLPLogsEndpointTmpl string = "%s/v1/log-streams/%s/logs"

	// auditCollectorEndpoint mirrors the receiver the bundled audit-export
	// collector binds to. Both sides are fixed: the collector is supervised by
	// this process and only ever listens on loopback.
	auditCollectorEndpoint string = "http://127.0.0.1:4318/v1/logs"
)

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

// newAuditCollectorProcessor ships to the audit-export collector bundled
// alongside the runner. The collector filters to records tagged
// nuon.audit="true" before forwarding to the customer's backend, so everything
// else sent over this hop is dropped there.
func newAuditCollectorProcessor() (log.Processor, error) {
	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpointURL(auditCollectorEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize audit collector log exporter: %w", err)
	}

	return newBatchProcessor(exporter), nil
}

func newBatchProcessor(exporter log.Exporter) log.Processor {
	return log.NewBatchProcessor(exporter,
		log.WithExportMaxBatchSize(25),
		log.WithExportInterval(time.Second*5))
}
