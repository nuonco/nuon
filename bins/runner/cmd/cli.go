package cmd

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/api"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/auth"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/componenthealth"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/drain"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/metrics"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/slog"
	"github.com/nuonco/nuon/bins/runner/internal/registry"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/log"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
	ociresolve "github.com/nuonco/nuon/pkg/runner/oci/resolve"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

type cli struct{}

func (c *cli) commonProviders() []fx.Option {
	// providers for both runner modes: mng and (org |install)
	return []fx.Option{
		fx.Provide(runnerconfig.NewConfig),
		fx.Provide(validator.New),
		// logging and error handling
		fx.Provide(slog.AsSystemProvider(slog.NewSystemProvider)),
		fx.Provide(log.AsSystemLogger(log.NewSystem)),
		fx.Provide(log.AsDevLogger(log.NewDev)),
		fx.WithLogger(log.NewFXLog),
		fx.Provide(errs.NewRecorder),
		// auth: fetch token via IMDS (or use existing token from env)
		fx.Provide(auth.New),
		// api client and settings (depend on auth token)
		fx.Provide(api.New),
		fx.Provide(settings.New),
		fx.Provide(heartbeater.New),
		fx.Provide(process.New),
		fx.Provide(process.NewShutdownPoller),
		fx.Provide(drain.New),
		fx.Provide(metrics.New),
		// shared cluster access + terraform state captured by deploy handlers for
		// the component-health engine
		fx.Provide(componenthealth.NewClusterProvider),
		fx.Provide(componenthealth.NewTerraformProvider),
		fx.Provide(componenthealth.NewManifestKindsProvider),
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
