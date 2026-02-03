package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerAuthAPI() error {
	cmd := &cobra.Command{
		Use:   "api-auth",
		Short: "run only the auth API",
		Run:   c.runAuthAPI,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runAuthAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	// Add API-specific modules - only auth API (includes auth service)
	providers = append(providers,
		fxmodules.MiddlewaresModule,
		fxmodules.AuthServicesModule,
		fxmodules.AuthAPIModule,
	)

	fx.New(providers...).Run()
}
