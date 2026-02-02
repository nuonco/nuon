package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerInternalAPI() error {
	cmd := &cobra.Command{
		Use:   "api-internal",
		Short: "run only the internal API",
		Run:   c.runInternalAPI,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runInternalAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	// Add API-specific modules - only internal API (excludes auth service)
	providers = append(providers,
		fxmodules.MiddlewaresModule,
		fxmodules.InternalServicesModule,
		fxmodules.InternalAPIModule,
	)

	fx.New(providers...).Run()
}
