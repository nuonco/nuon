package cmd

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/api"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/heartbeater"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/metrics"
	ocicopy "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/copy"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/slog"
	"github.com/powertoolsdev/mono/bins/runner/internal/registry"
)

type cli struct{}

func (c *cli) commonProviders() []fx.Option {
	// providers for both runner modes: mng and (org |install)
	return []fx.Option{
		fx.Provide(settings.New),
		fx.Provide(internal.NewConfig),
		fx.Provide(validator.New),
		fx.Provide(api.New),
		fx.Provide(heartbeater.New),
		fx.Provide(metrics.New),
		// logging and error handling
		fx.Provide(slog.AsSystemProvider(slog.NewSystemProvider)),
		fx.Provide(log.AsSystemLogger(log.NewSystem)),
		fx.Provide(log.AsDevLogger(log.NewDev)),
		fx.Provide(errs.NewRecorder),
	}
}

func (c *cli) providers() []fx.Option {
	// providers for (org |install) mode
	return append(
		c.commonProviders(),
		[]fx.Option{
			fx.Provide(ocicopy.New),
			fx.Provide(registry.New),

			// NOTE(jm): we plan to deprecate the default loggers, so each logger is forced to be depended on via
			// name.
			fx.Provide(log.NewSystem),
		}...,
	)
}
