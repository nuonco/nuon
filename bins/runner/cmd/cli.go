package cmd

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/api"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/heartbeater"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	ocicopy "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/copy"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/runnerctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

type cli struct{}

func (c *cli) providers() []fx.Option {
	return []fx.Option{
		fx.Provide(runnerctx.New),
		fx.Provide(settings.New),
		fx.Provide(internal.NewConfig),
		fx.Provide(validator.New),
		fx.Provide(api.New),
		fx.Provide(heartbeater.New),
		fx.Provide(log.New),
		fx.Provide(log.NewHclog),
		fx.Provide(errs.NewRecorder),
		fx.Provide(ocicopy.New),
	}
}
