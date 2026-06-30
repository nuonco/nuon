package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerMCPAPI() error {
	cmd := &cobra.Command{
		Use:   "api-mcp",
		Short: "run only the MCP server (streamable HTTP)",
		Run:   c.runMCPAPI,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runMCPAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	providers = append(providers,
		fxmodules.MiddlewaresModule,
		fxmodules.AllServicesModule,
		fxmodules.MCPAPIModule,
	)

	fx.New(providers...).Run()
}
