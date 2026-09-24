package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

func (c *cli) registerNuonctlMCPAPI() error {
	cmd := &cobra.Command{
		Use:   "nuonctl-mcp-api",
		Short: "run the employee-only nuonctl MCP server",
		Run:   c.runNuonctlMCPAPI,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runNuonctlMCPAPI(cmd *cobra.Command, _ []string) {
	providers := make([]fx.Option, 0)
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	providers = append(providers,
		fxmodules.NuonctlMCPServicesModule,
		fxmodules.NuonctlMCPAPIModule,
	)

	fx.New(providers...).Run()
}
