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
		Long: `Run a Model Context Protocol server over stdio, exposing Nuon operations as
tools for LLM clients like Claude Code and Claude Desktop.

Read-only by default. Pass --allow-writes to also expose mutating tools
(create_install, deploy_component).

Example Claude Code config (.mcp.json):

  {"mcpServers": {"nuon": {"command": "nuon", "args": ["mcp"]}}}`,
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
		Annotations:       outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if ReadOnly || readOnlyFromEnv() {
				allowWrites = false
			}
			svc := mcpserver.New(c.apiClient, c.cfg, allowWrites)
			return svc.Run(cmd.Context())
		}),
	}
	mcpCmd.Flags().BoolVar(&allowWrites, "allow-writes", false, "expose mutating tools (create_install, deploy_component)")

	return mcpCmd
}
