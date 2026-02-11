package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerAdminDashboardAPI() error {
	cmd := &cobra.Command{
		Use:   "api-admin",
		Short: "run only the admin dashboard API",
		Run:   c.runAdminDashboardAPI,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runAdminDashboardAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	providers = append(providers,
		fxmodules.MiddlewaresModule,
		fxmodules.AdminDashboardServicesModule,
		fxmodules.AdminDashboardAPIModule,
	)

	fx.New(providers...).Run()
}
