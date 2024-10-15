package slog

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

const (
	defaultOTLPLogsEndpointTmpl string = "%s/v1/runners/%s/logs"
)

type Params struct {
	fx.In

	Cfg      *internal.Config
	LC       fx.Lifecycle
	Settings *settings.Settings

	Provider *log.LoggerProvider `name:"system"`
}

func NewOTELProvider(params Params) (*log.LoggerProvider, error) {
	if !params.Settings.EnableLogging {
		return log.NewLoggerProvider(), nil
	}

	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	url := fmt.Sprintf(defaultOTLPLogsEndpointTmpl, params.Cfg.RunnerAPIURL, params.Cfg.RunnerID)

	// TODO(jm): do this configuration in a less hacky way
	os.Setenv("OTEL_EXPORTER_OTLP_ENCODING", "json")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", url)
	os.Setenv("OTEL_SERVICE_NAME", "runner")
	os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer "+params.Cfg.RunnerAPIToken)
	// os.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
	// os.Setenv("OTEL_EXPORTER_OTLP_LOGS_COMPRESSION", "none")

	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("unable to initialize otlp log exporter: %w", err)
	}

	rsrc := getResource(params.Settings)
	// Create the logger provider
	lp := log.NewLoggerProvider(
		log.WithResource(rsrc),
		log.WithProcessor(
			log.NewBatchProcessor(logExporter),
		),
	)
	params.LC.Append(lifecycleHook(cancelFn, lp))

	return lp, nil
}
