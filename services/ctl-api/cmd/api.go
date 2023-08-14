package cmd

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/docs"
	appshooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/hooks"
	appsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/service"
	generalservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/service"
	installshooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	installsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/service"
	orgshooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/hooks"
	orgsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/service"
	sandboxesservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/sandboxes/service"
	vcsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/service"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/health"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/metrics"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func (c *cli) registerAPI() error {
	var runApiCmd = &cobra.Command{
		Use:   "api",
		Short: "run api",
		Run:   c.runAPI,
	}
	rootCmd.AddCommand(runApiCmd)
	return nil
}

func (c *cli) runAPI(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{
		// add app hooks
		fx.Provide(appshooks.New),
		fx.Provide(installshooks.New),
		fx.Provide(orgshooks.New),

		// add middlewares
		fx.Provide(api.AsMiddleware(metrics.New)),

		// add endpoints
		fx.Provide(api.AsService(docs.New)),
		fx.Provide(api.AsService(health.New)),
		fx.Provide(api.AsService(orgsservice.New)),
		fx.Provide(api.AsService(appsservice.New)),
		fx.Provide(api.AsService(vcsservice.New)),
		fx.Provide(api.AsService(generalservice.New)),
		fx.Provide(api.AsService(sandboxesservice.New)),
		fx.Provide(api.AsService(installsservice.New)),

		fx.Provide(fx.Annotate(api.NewAPI, fx.ParamTags(`group:"services"`, `group:"middlewares"`))),
		fx.Invoke(func(*api.API) {

		}),
	}

	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
