// Package logstream wires a *slog.Logger to either stdout or the Nuon
// ctl-api OTLP log-stream ingest endpoint.
//
// The OTLP path mirrors bins/runner/internal/pkg/slog/otel.go: batched
// otlploghttp exporter posting to {RunnerAPIURL}/v1/log-streams/{id}/logs
// with a Bearer write token.
package logstream

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

const otlpEndpointTmpl = "%s/v1/log-streams/%s/logs"

// Config configures an OTLP-backed slog logger.
type Config struct {
	RunnerAPIURL string
	LogStreamID  string
	WriteToken   string
	// ServiceName is set on the OTEL resource (e.g. "stack").
	ServiceName string
	// Attrs are added as resource attributes (e.g. install_id).
	Attrs map[string]string
}

// Provider holds an OTEL log provider that must be shut down when done.
type Provider struct {
	lp     *sdklog.LoggerProvider
	logger *slog.Logger
}

func (p *Provider) Logger() *slog.Logger { return p.logger }

func (p *Provider) Shutdown(ctx context.Context) error {
	if p.lp == nil {
		return nil
	}
	return p.lp.Shutdown(ctx)
}

// NewStdout returns a Provider that writes to stdout. Useful for local dev.
func NewStdout(serviceName string) *Provider {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	return &Provider{logger: slog.New(h).With("service", serviceName)}
}

// New creates an OTLP-backed Provider. Caller must call Shutdown.
func New(ctx context.Context, cfg Config) (*Provider, error) {
	if cfg.RunnerAPIURL == "" || cfg.LogStreamID == "" || cfg.WriteToken == "" {
		return nil, fmt.Errorf("logstream: RunnerAPIURL, LogStreamID, WriteToken all required")
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "stack"
	}

	endpoint := fmt.Sprintf(otlpEndpointTmpl, cfg.RunnerAPIURL, cfg.LogStreamID)
	exp, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(endpoint),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + cfg.WriteToken,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("logstream: init otlp exporter: %w", err)
	}

	attrs := []attribute.KeyValue{attribute.String("service.name", cfg.ServiceName)}
	for k, v := range cfg.Attrs {
		attrs = append(attrs, attribute.String(k, v))
	}
	rsrc := resource.NewSchemaless(attrs...)

	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(rsrc),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exp,
			sdklog.WithExportMaxBatchSize(25),
			sdklog.WithExportInterval(5*time.Second),
		)),
	)

	logger := otelslog.NewLogger(cfg.ServiceName, otelslog.WithLoggerProvider(lp))
	return &Provider{lp: lp, logger: logger}, nil
}
