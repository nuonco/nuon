package slog

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

const (
	defaultOTLPLogsEndpointTmpl string = "%s/v1/log-streams/%s/logs"
)

func NewOTELProvider(cfg *internal.Config, set *settings.Settings, logStreamID string) (*log.LoggerProvider, error) {
	if !set.EnableLogging {
		return log.NewLoggerProvider(), nil
	}

	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	url := fmt.Sprintf(defaultOTLPLogsEndpointTmpl, cfg.RunnerAPIURL, logStreamID)
	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(url),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + cfg.RunnerAPIToken,
		}),
	)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("unable to initialize otlp log exporter: %w", err)
	}

	rsrc := getResource(set)
	// Create the logger provider
	lp := log.NewLoggerProvider(
		log.WithResource(rsrc),
		log.WithProcessor(
			log.NewBatchProcessor(logExporter,
				log.WithExportMaxBatchSize(25),
				log.WithExportInterval(time.Second*5)),
		),
	)

	return lp, nil
}
