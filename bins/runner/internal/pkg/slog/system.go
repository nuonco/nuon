package slog

import (
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
)

type SystemParams struct {
	fx.In
}

func NewSystemProvider(params SystemParams) (*log.LoggerProvider, error) {
	lp := log.NewLoggerProvider()

	return lp, nil
}
