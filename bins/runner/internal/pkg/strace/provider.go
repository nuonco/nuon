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

func NewProcessProvider(cfg *runnerconfig.Config, set *settings.Settings) (*sdktrace.TracerProvider, error) {
	if !set.EnableLogging {
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
