package cmd

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/db/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/protos"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/terraformcloud"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/validator"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/waypoint"
	appshooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/hooks"
	componentsshooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/hooks"
	installshooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	orgshooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/hooks"
	releaseshooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/hooks"
	"go.uber.org/fx"
)

type cli struct{}

func (c *cli) providers() []fx.Option {
	return []fx.Option{
		fx.Provide(internal.NewConfig),

		// various dependencies
		fx.Provide(log.New),
		fx.Provide(github.New),
		fx.Provide(metrics.New),
		fx.Provide(migrations.New),
		fx.Provide(db.New),
		fx.Provide(temporal.New),
		fx.Provide(validator.New),
		fx.Provide(protos.New),
		fx.Provide(terraformcloud.NewTerraformCloud),
		fx.Provide(terraformcloud.NewOrgsOutputs),
		fx.Provide(waypoint.New),

		// add app hooks
		fx.Provide(appshooks.New),
		fx.Provide(installshooks.New),
		fx.Provide(orgshooks.New),
		fx.Provide(componentsshooks.New),
		fx.Provide(releaseshooks.New),
	}
}
