package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerSlackAPI() error {
	cmd := &cobra.Command{
		Use:   "api-slack",
		Short: "run only the Slack API (OAuth callback, slash commands, events webhooks)",
		Run:   c.runSlackAPI,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runSlackAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	providers = append(providers,
		fxmodules.MiddlewaresModule,
		fxmodules.SlackServicesModule,
		fxmodules.SlackAPIModule,
	)

	fx.New(providers...).Run()
}
