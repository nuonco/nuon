package cmd

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/api"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/auth"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/errs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/log"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/metrics"
	ocicopy "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/copy"
	ociresolve "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/resolve"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/slog"
	"github.com/nuonco/nuon/bins/runner/internal/registry"
)

type cli struct{}

func (c *cli) commonProviders() []fx.Option {
	// providers for both runner modes: mng and (org |install)
	return []fx.Option{
		fx.Provide(internal.NewConfig),
		fx.Provide(validator.New),
		// logging and error handling
		fx.Provide(slog.AsSystemProvider(slog.NewSystemProvider)),
		fx.Provide(log.AsSystemLogger(log.NewSystem)),
		fx.Provide(log.AsDevLogger(log.NewDev)),
		fx.Provide(errs.NewRecorder),
		// auth: fetch token via IMDS (or use existing token from env)
		fx.Provide(auth.New),
		// api client and settings (depend on auth token)
		fx.Provide(api.New),
		fx.Provide(settings.New),
		fx.Provide(heartbeater.New),
		fx.Provide(process.New),
		fx.Provide(process.NewShutdownPoller),
		fx.Provide(metrics.New),
	}
}

func (c *cli) providers() []fx.Option {
	// providers for (org |install) mode
	return append(
		c.commonProviders(),
		[]fx.Option{
			fx.Provide(ocicopy.New),
			fx.Provide(ociresolve.New),
			fx.Provide(registry.New),

			// NOTE(jm): we plan to deprecate the default loggers, so each logger is forced to be depended on via
			// name.
			fx.Provide(log.NewSystem),
		}...,
	)
}
