package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/mcpserver"
)

func (c *cli) mcpCmd() *cobra.Command {
	var allowWrites bool

	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Deprecated: use `nuon agents mcp`",
		Long: `Deprecated: use "nuon agents mcp" instead.

Run a Model Context Protocol server over stdio that proxies to the Nuon
control plane MCP server. Tools are discovered from the upstream server
and forwarded transparently.

Read-only by default. Pass --allow-writes to also expose mutating tools.

Example Claude Code config (.mcp.json):

  {"mcpServers": {"nuon": {"command": "nuon", "args": ["agents", "mcp"]}}}`,
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
		Annotations:       outputsAnnotation(OutputTable),
		Deprecated:        "use \"nuon agents mcp\" instead",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(os.Stderr, `Warning: "nuon mcp" is deprecated; use "nuon agents mcp" instead.`)

			// An MCP client drives this over piped stdio, so keep proxying for
			// existing .mcp.json configs. A human on a TTY would just see the
			// server block on stdin, so point them at the new command instead.
			if c.cfg.Interactive {
				fmt.Fprintln(os.Stderr, `Run "nuon agents mcp" to start the stdio MCP proxy.`)
				return nil
			}

			if ReadOnly || readOnlyFromEnv() {
				allowWrites = false
			}
			return mcpserver.New(c.cfg, allowWrites).Run(cmd.Context())
		}),
	}
	mcpCmd.Flags().BoolVar(&allowWrites, "allow-writes", false, "expose mutating tools whose descriptions start with WRITE OPERATION:")

	mcpCmd.AddCommand(c.mcpSetupCmd())

	return mcpCmd
}
