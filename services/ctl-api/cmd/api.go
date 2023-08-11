package cmd

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/api"
	orgsservice "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/service"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/health"
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
		// add endpoints
		fx.Provide(api.AsService(health.New)),
		fx.Provide(api.AsService(orgsservice.New)),
		fx.Provide(fx.Annotate(api.NewAPI, fx.ParamTags(`group:"services"`))),
		fx.Invoke(func(*api.API) {
		}),
	}

	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
