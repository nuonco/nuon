package cmd

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	componentshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	installshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/psql"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/loops"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/terraformcloud"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/waypoint"
)

type cli struct{}

func (c *cli) providers() []fx.Option {
	return []fx.Option{
		fx.Provide(internal.NewConfig),

		// various dependencies
		fx.Provide(log.New),
		fx.Provide(dblog.New),
		fx.Provide(loops.New),
		fx.Provide(github.New),
		fx.Provide(metrics.New),
		fx.Provide(migrations.New),
		fx.Provide(db.AsPSQL(psql.New)),
		fx.Provide(db.AsCH(ch.New)),
		fx.Provide(temporal.New),
		fx.Provide(validator.New),
		fx.Provide(protos.New),
		fx.Provide(terraformcloud.NewTerraformCloud),
		fx.Provide(terraformcloud.NewOrgsOutputs),
		fx.Provide(waypoint.New),
		fx.Provide(activities.New),
		fx.Provide(notifications.New),
		fx.Provide(eventloop.New),
		fx.Provide(teventloop.New),
		fx.Provide(authz.New),
		fx.Provide(account.New),
		fx.Provide(analytics.NewContextWriter),
		fx.Provide(analytics.NewTemporalWriter),

		// add helpers for each domain
		fx.Provide(vcshelpers.New),
		fx.Provide(componentshelpers.New),
		fx.Provide(appshelpers.New),
		fx.Provide(installshelpers.New),
		fx.Provide(runnershelpers.New),
	}
}
