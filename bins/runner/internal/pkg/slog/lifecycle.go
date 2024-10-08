package slog

import (
	"context"

	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
)

func lifecycleHook(cancelFn func(), lp *log.LoggerProvider) fx.Hook {
	return fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			lp.Shutdown(ctx)
			cancelFn()
			return nil
		},
	}
}
