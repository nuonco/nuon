package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/profiles"
	actionsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/service"
	appsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/service"
	componentsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/service"
	generalservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/service"
	installersservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installers/service"
	installsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/service"
	orgsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/service"
	releasesservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/service"
	runnersservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/service"
	vcsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/service"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/health"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/httpbin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/admin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/config"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/cors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/global"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/headers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/invites"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/pagination"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/panicker"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/patcher"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/public"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/size"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/timeout"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/docs"
)

func (c *cli) registerAPI() error {
	runApiCmd := &cobra.Command{
		Use:   "api",
		Short: "run api",
		Run:   c.runAPI,
	}
	rootCmd.AddCommand(runApiCmd)
	return nil
}

func (c *cli) runAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)
	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))
	providers = append(providers,
		// add middlewares
		fx.Provide(middlewares.AsMiddleware(stderr.New)),
		fx.Provide(middlewares.AsMiddleware(global.New)),
		fx.Provide(middlewares.AsMiddleware(metrics.New)),
		fx.Provide(middlewares.AsMiddleware(metrics.NewInternal)),
		fx.Provide(middlewares.AsMiddleware(metrics.NewRunner)),
		fx.Provide(middlewares.AsMiddleware(headers.New)),
		fx.Provide(middlewares.AsMiddleware(auth.New)),
		fx.Provide(middlewares.AsMiddleware(org.New)),
		fx.Provide(middlewares.AsMiddleware(org.NewRunner)),
		fx.Provide(middlewares.AsMiddleware(public.New)),
		fx.Provide(middlewares.AsMiddleware(pagination.New)),
		fx.Provide(middlewares.AsMiddleware(cors.New)),
		fx.Provide(middlewares.AsMiddleware(config.New)),
		fx.Provide(middlewares.AsMiddleware(patcher.New)),
		fx.Provide(middlewares.AsMiddleware(invites.New)),
		fx.Provide(middlewares.AsMiddleware(admin.New)),
		fx.Provide(middlewares.AsMiddleware(log.New)),
		fx.Provide(middlewares.AsMiddleware(log.New)),
		fx.Provide(middlewares.AsMiddleware(size.New)),
		fx.Provide(middlewares.AsMiddleware(timeout.New)),
		fx.Provide(middlewares.AsMiddleware(panicker.New)),

		// add endpoints
		fx.Provide(api.AsService(docs.New)),
		fx.Provide(api.AsService(health.New)),
		fx.Provide(api.AsService(orgsservice.New)),
		fx.Provide(api.AsService(appsservice.New)),
		fx.Provide(api.AsService(vcsservice.New)),
		fx.Provide(api.AsService(generalservice.New)),
		fx.Provide(api.AsService(installsservice.New)),
		fx.Provide(api.AsService(installersservice.New)),
		fx.Provide(api.AsService(componentsservice.New)),
		fx.Provide(api.AsService(runnersservice.New)),
		fx.Provide(api.AsService(releasesservice.New)),
		fx.Provide(api.AsService(actionsservice.New)),
		fx.Provide(api.AsService(httpbin.New)),

		// add api
		fx.Provide(api.AsAPI(api.NewPublicAPI)),
		fx.Provide(api.AsAPI(api.NewRunnerAPI)),
		fx.Provide(api.AsAPI(api.NewInternalAPI)),

		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(api.APIGroupParam(func([]*api.API) {})),
	)

	fx.New(providers...).Run()
}
