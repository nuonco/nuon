package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gorm.io/gorm"

	appsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/service"
	componentsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/service"
	generalservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/service"
	installersservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installers/service"
	installsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/service"
	orgsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/service"
	releasesservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/service"
	vcsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/service"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/health"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/config"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/cors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/global"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/headers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/invites"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/public"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
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
	providers := []fx.Option{
		// add middlewares
		fx.Provide(api.AsMiddleware(stderr.New)),
		fx.Provide(api.AsMiddleware(global.New)),
		fx.Provide(api.AsMiddleware(metrics.New)),
		fx.Provide(api.AsMiddleware(metrics.NewInternal)),
		fx.Provide(api.AsMiddleware(headers.New)),
		fx.Provide(api.AsMiddleware(auth.New)),
		fx.Provide(api.AsMiddleware(org.New)),
		fx.Provide(api.AsMiddleware(public.New)),
		fx.Provide(api.AsMiddleware(cors.New)),
		fx.Provide(api.AsMiddleware(config.New)),
		fx.Provide(api.AsMiddleware(invites.New)),

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
		fx.Provide(api.AsService(releasesservice.New)),

		fx.Provide(fx.Annotate(api.NewAPI, fx.ParamTags(`group:"services"`, `group:"middlewares"`))),
		fx.Invoke(func(*gorm.DB) {}),
		fx.Invoke(func(*api.API) {}),
	}

	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
