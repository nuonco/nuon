package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerRunnerAPI() error {
	cmd := &cobra.Command{
		Use:   "api-runner",
		Short: "run only the runner API",
		Run:   c.runRunnerAPI,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runRunnerAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	// Add API-specific modules - only runner API (excludes auth service)
	providers = append(providers,
		fxmodules.MiddlewaresModule,
		fxmodules.RunnerServicesModule,
		fxmodules.RunnerAPIModule,
	)

	fx.New(providers...).Run()
}
