package cmd

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/temporal"
	"go.uber.org/fx"
)

type cli struct{}

func (c *cli) providers() []fx.Option {
	return []fx.Option{
		fx.Provide(internal.NewConfig),
		fx.Provide(log.New),
		fx.Provide(github.NewGithubClient),
		fx.Provide(metrics.NewWriter),
		fx.Provide(db.New),
		fx.Provide(temporal.New),
		fx.Provide(validator.New),
	}
}
