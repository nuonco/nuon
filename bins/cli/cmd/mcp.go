package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/mcpserver"
)

func (c *cli) mcpCmd() *cobra.Command {
	var allowWrites bool

	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run an MCP server exposing Nuon tools",
		Long: `Run a Model Context Protocol server over stdio that proxies to the Nuon
control plane MCP server. Tools are discovered from the upstream server
and forwarded transparently.

Read-only by default. Pass --allow-writes to also expose mutating tools.

Example Claude Code config (.mcp.json):

  {"mcpServers": {"nuon": {"command": "nuon", "args": ["mcp"]}}}`,
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
		Annotations:       outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if ReadOnly || readOnlyFromEnv() {
				allowWrites = false
			}
			svc := mcpserver.New(c.cfg, allowWrites)
			return svc.Run(cmd.Context())
		}),
	}
	mcpCmd.Flags().BoolVar(&allowWrites, "allow-writes", false, "expose mutating tools (create_install, deploy_component)")

	mcpCmd.AddCommand(c.mcpSetupCmd())

	return mcpCmd
}
