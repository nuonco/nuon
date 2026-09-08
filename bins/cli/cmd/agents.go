package cmd

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/mcpserver"
)

func (c *cli) agentsCmd() *cobra.Command {
	agentsCmd := &cobra.Command{
		Use:   "agents",
		Short: "Agent-facing helpers for driving Nuon with LLMs",
		Long: `Commands for LLM agents working with Nuon.

Start here:
  nuon agents context   Print markdown orientation (auth, selection, how to use MCP)
  nuon agents mcp       Run a local stdio MCP proxy that injects your API token and org ID

Register with an MCP client:

  claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes
  amp mcp add nuon -- nuon agents mcp --allow-writes

Cursor / Cursor Agent: write ~/.cursor/mcp.json then agent mcp enable nuon.

Direct HTTP: point an MCP client at the URL from "nuon agents context"
with Bearer token and X-Nuon-Org-ID headers.`,
		GroupID:     AdditionalGroup.ID,
		Annotations: annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		}),
	}

	agentsCmd.AddCommand(c.agentsContextCmd())
	agentsCmd.AddCommand(c.agentsMCPCmd())

	return agentsCmd
}

func (c *cli) agentsContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "context",
		Short: "Print agent orientation markdown for the current CLI context",
		Long: `Print a large markdown document describing the current Nuon CLI context
and how an agent should interact with Nuon (local MCP proxy vs HTTP API MCP).

Intended for LLM agents: run this first when asked to work with Nuon.`,
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		Run: c.wrapCmd(func(_ *cobra.Command, _ []string) error {
			fmt.Print(c.agentsContextMarkdown())
			return nil
		}),
	}
}

func (c *cli) agentsMCPCmd() *cobra.Command {
	var allowWrites bool
	var mcpURL string
	var serverName string

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run a local stdio MCP proxy to the Nuon control plane",
		Long: `Run a Model Context Protocol server over stdio that proxies to the Nuon
control plane MCP server. Tools are discovered from the upstream server
and forwarded transparently.

Injects Authorization (Bearer) and X-Nuon-Org-ID from ~/.nuon on every
upstream request. Read-only by default; pass --allow-writes to also
expose mutating tools (descriptions prefixed with "WRITE OPERATION:").

Example — register:

  claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes
  amp mcp add nuon -- nuon agents mcp --allow-writes
  # Cursor: write ~/.cursor/mcp.json, then agent mcp enable nuon

Or add to .mcp.json:

  {"mcpServers": {"nuon": {"command": "nuon", "args": ["agents", "mcp", "--allow-writes"]}}}

Example — override the upstream server:

  nuon agents mcp --allow-writes --url https://mcp.example.com/mcp --name nuon-example

Run "nuon agents context" to see which MCP URL resolves from your config.`,
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if ReadOnly || readOnlyFromEnv() {
				allowWrites = false
			}
			opts := make([]mcpserver.Option, 0, 2)
			if mcpURL != "" {
				opts = append(opts, mcpserver.WithEndpoint(mcpURL))
			}
			if serverName != "" {
				opts = append(opts, mcpserver.WithName(serverName))
			}
			return mcpserver.New(c.cfg, allowWrites, opts...).Run(cmd.Context())
		}),
	}
	cmd.Flags().BoolVar(&allowWrites, "allow-writes", false, "expose mutating tools whose descriptions start with WRITE OPERATION:")
	cmd.Flags().StringVar(&mcpURL, "url", "", "upstream MCP server URL (derived from api.<hostname>, or localhost; otherwise required)")
	cmd.Flags().StringVar(&serverName, "name", "", "MCP server name exposed to the client (default nuon, derived from the configured API URL)")

	return cmd
}

//go:embed agents_context.md
var agentsContextDoc string

type agentsContext struct {
	Authed    string
	APIURL    string
	MCPURL    string
	OrgID     string
	AppID     string
	InstallID string
}

func (c *cli) agentsContextMarkdown() string {
	cfg := c.cfg

	data := agentsContext{
		Authed:    "no",
		APIURL:    "(unset)",
		MCPURL:    "(unknown — config not loaded)",
		OrgID:     "(none selected)",
		AppID:     "(none selected)",
		InstallID: "(none selected)",
	}
	if cfg != nil {
		data.APIURL = cfg.APIURL
		if derived, err := mcpserver.EndpointFromAPIURL(cfg.APIURL); err != nil {
			data.MCPURL = err.Error()
		} else {
			data.MCPURL = derived
		}
		if cfg.OrgID != "" {
			data.OrgID = cfg.OrgID
		}
		if cfg.AppID != "" {
			data.AppID = cfg.AppID
		}
		if cfg.InstallID != "" {
			data.InstallID = cfg.InstallID
		}
		if cfg.APIToken != "" {
			data.Authed = "yes (API token present in ~/.nuon)"
		}
	}

	tmpl, err := template.New("agents_context").Parse(agentsContextDoc)
	if err != nil {
		return agentsContextDoc
	}

	var b strings.Builder
	if err := tmpl.Execute(&b, data); err != nil {
		return agentsContextDoc
	}

	return b.String()
}
