package slog

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/otelresource"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

const (
	defaultOTLPLogsEndpointTmpl string = "%s/v1/log-streams/%s/logs"
	otlpExportTimeout                  = 10 * time.Second
)

func NewOTELProvider(cfg *runnerconfig.Config, set *settings.Settings, logStreamID string) (*log.LoggerProvider, error) {
	opts := []log.LoggerProviderOption{
		log.WithResource(otelresource.New(set, logStreamID)),
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
		otlploghttp.WithHTTPClient(&http.Client{
			Timeout: otlpExportTimeout,
			Transport: &bearerTransport{
				token: cfg.APIToken,
				next:  http.DefaultTransport,
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize otlp log exporter: %w", err)
	}

	return newBatchProcessor(exporter), nil
}

type bearerTransport struct {
	token func() string
	next  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token())
	return t.next.RoundTrip(req)
}

func newBatchProcessor(exporter log.Exporter) log.Processor {
	return log.NewBatchProcessor(exporter,
		log.WithExportMaxBatchSize(25),
		log.WithExportInterval(time.Second*5))
}
