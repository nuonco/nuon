package runnerctx

import (
	"context"

	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal"
)

func New(cfg *internal.Config, lc fx.Lifecycle) context.Context {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, runnerID{}, cfg.RunnerID)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancelFn()
			return nil
		},
	})

	return ctx
}
