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

func newBatchProcessor(exporter log.Exporter) log.Processor {
	return log.NewBatchProcessor(exporter,
		log.WithExportMaxBatchSize(25),
		log.WithExportInterval(time.Second*5))
}
