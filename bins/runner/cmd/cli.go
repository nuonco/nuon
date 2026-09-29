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

type cli struct {
	extraProviders []fx.Option
}

func (c *cli) commonProviders() []fx.Option {
	return []fx.Option{
		fx.Provide(runnerconfig.NewConfig),
		fx.Provide(validator.New),
		fx.Provide(slog.AsSystemProvider(slog.NewSystemProvider)),
		fx.Provide(log.AsSystemLogger(log.NewSystem)),
		fx.Provide(log.AsDevLogger(log.NewDev)),
		fx.WithLogger(log.NewFXLog),
		fx.Provide(errs.NewRecorder),
		fx.Provide(auth.New),
		fx.Provide(api.New),
		fx.Provide(settings.New),
		fx.Provide(heartbeater.New),
		fx.Provide(process.New),
		fx.Provide(process.NewShutdownPoller),
		fx.Provide(drain.New),
		fx.Provide(metrics.New),
		fx.Provide(componenthealth.NewClusterProvider),
		fx.Provide(componenthealth.NewTerraformProvider),
		fx.Provide(componenthealth.NewManifestKindsProvider),
	}
}

func (c *cli) providers() []fx.Option {
	return append(
		c.commonProviders(),
		[]fx.Option{
			fx.Provide(ocicopy.New),
			fx.Provide(ociresolve.New),
			fx.Provide(registry.New),

			fx.Provide(log.NewSystem),
		}...,
	)
}
