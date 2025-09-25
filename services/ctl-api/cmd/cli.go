package cmd

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	actionshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	componentshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	installshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	orgshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cloudformation"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/psql"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/features"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/loops"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	signaldb "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
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
		fx.Provide(propagator.New),
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
		fx.Provide(signaldb.NewPayloadConverter),
		fx.Provide(dataconverter.New),
		fx.Provide(temporal.New),
		fx.Provide(validator.New),
		fx.Provide(notifications.New),
		fx.Provide(eventloop.New),
		fx.Provide(teventloop.New),
		fx.Provide(authz.New),
		fx.Provide(features.New),
		fx.Provide(account.New),
		fx.Provide(analytics.New),
		fx.Provide(analytics.NewTemporal),
		fx.Provide(cloudformation.NewTemplates),

		// add helpers for each domain
		fx.Provide(vcshelpers.New),
		fx.Provide(actionshelpers.New),
		fx.Provide(componentshelpers.New),
		fx.Provide(orgshelpers.New),
		fx.Provide(appshelpers.New),
		fx.Provide(installshelpers.New),
		fx.Provide(runnershelpers.New),
	}
}
