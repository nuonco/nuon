package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"

	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerAPI() error {
	runApiCmd := &cobra.Command{
		Use:   "api",
		Short: "run all APIs (public, internal, runner, auth, admin-dashboard, slack)",
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

	// Add API-specific modules - all APIs (includes auth service)
	providers = append(providers,
		fxmodules.MiddlewaresModule,
		fxmodules.AllServicesModule,
		fxmodules.AllAPIsModule,
	)

	fx.New(providers...).Run()
}
