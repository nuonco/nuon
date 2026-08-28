package strace

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

// NewProcessProvider constructs a process-scoped TracerProvider for the
// runner. Spans emitted via op.Start during a job execution are attributed to
// this provider and exported as OTLP JSON to the runner traces ingest
// endpoint.
//
// Process scope (one provider per runner process) is intentional and mirrors
// how slog.NewOTELProvider creates a per-job log provider: traces correlate
// across the full process lifetime via runner_id, while logs correlate per
// log_stream_id. The two share only the Resource builder.
func NewProcessProvider(cfg *runnerconfig.Config, set *settings.Settings) (*sdktrace.TracerProvider, error) {
	// Offline customer-managed logging uses a loopback OTLP endpoint, but traces still have no
	// offline sink. A no-op provider prevents them from reaching the local API.
	if !set.EnableLogging || cfg.OTLPLogsEndpoint != "" {
		return sdktrace.NewTracerProvider(), nil
	}

	exp, err := otlptrace.New(context.Background(), newJSONHTTPClient(cfg.RunnerAPIURL, cfg.RunnerID, cfg.RunnerAPIToken))
	if err != nil {
		return nil, fmt.Errorf("strace: init otlp trace exporter: %w", err)
	}

	rsrc := getResource(set)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(rsrc),
		sdktrace.WithBatcher(exp,
			sdktrace.WithMaxExportBatchSize(64),
			sdktrace.WithBatchTimeout(2*time.Second),
		),
	)
	return tp, nil
}
