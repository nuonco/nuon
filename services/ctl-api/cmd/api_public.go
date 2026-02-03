package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerPublicAPI() error {
	cmd := &cobra.Command{
		Use:   "api-public",
		Short: "run only the public API",
		Run:   c.runPublicAPI,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runPublicAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	// Add API-specific modules - only public API (excludes auth service)
	providers = append(providers,
		fxmodules.MiddlewaresModule,
		fxmodules.PublicServicesModule,
		fxmodules.PublicAPIModule,
	)

	fx.New(providers...).Run()
}
