package slog

import (
	"fmt"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
)

type SystemParams struct {
	fx.In

	// ProcessLogProvider is the OTEL log provider from the process log stream.
	// When available, system logs are sent to the process log stream.
	// When nil, falls back to stdout.
	ProcessLogProvider *log.LoggerProvider `name:"process-log-provider" optional:"true"`
}

// NewSystemProvider returns the process log stream provider if available,
// falling back to stdout.
func NewSystemProvider(params SystemParams) (*log.LoggerProvider, error) {
	if params.ProcessLogProvider != nil {
		return params.ProcessLogProvider, nil
	}

	exporter, err := stdoutlog.New()
	if err != nil {
		return nil, fmt.Errorf("unable to create stdout log exporter: %w", err)
	}

	lp := log.NewLoggerProvider(
		log.WithProcessor(log.NewSimpleProcessor(exporter)),
	)

	return lp, nil
}
