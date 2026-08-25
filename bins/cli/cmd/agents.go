package cmd

import (
	"fmt"
	"strings"

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

Preferred MCP client config (Claude Code / Cursor):

  {"mcpServers": {"nuon": {"command": "nuon", "args": ["agents", "mcp"]}}}

You can also point an MCP client directly at the control-plane HTTP MCP
endpoint (see "nuon agents context"). Auth is Bearer token + optional
X-Nuon-Org-ID.

Coming later: nuon agents skills (scaffolding for apps, installs, and more).`,
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

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run a local stdio MCP proxy to the Nuon control plane",
		Long: `Run a Model Context Protocol server over stdio that proxies to the Nuon
control plane MCP server. Tools are discovered from the upstream server
and forwarded transparently.

Injects Authorization (Bearer) and X-Nuon-Org-ID from ~/.nuon on every
upstream request. Read-only by default; pass --allow-writes to also
expose mutating tools (descriptions prefixed with "WRITE OPERATION:").

Example Claude Code config (.mcp.json):

  {"mcpServers": {"nuon": {"command": "nuon", "args": ["agents", "mcp"]}}}`,
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if ReadOnly || readOnlyFromEnv() {
				allowWrites = false
			}
			return mcpserver.New(c.cfg, allowWrites).Run(cmd.Context())
		}),
	}
	cmd.Flags().BoolVar(&allowWrites, "allow-writes", false, "expose mutating tools whose descriptions start with WRITE OPERATION:")

	return cmd
}

func (c *cli) agentsContextMarkdown() string {
	cfg := c.cfg
	var b strings.Builder

	mcpURL := "(unknown — config not loaded)"
	apiURL := "(unset)"
	orgID := "(none selected)"
	appID := "(none selected)"
	installID := "(none selected)"
	authed := "no"
	if cfg != nil {
		apiURL = cfg.APIURL
		mcpURL = mcpserver.EndpointFromAPIURL(cfg.APIURL)
		if cfg.OrgID != "" {
			orgID = cfg.OrgID
		}
		if cfg.AppID != "" {
			appID = cfg.AppID
		}
		if cfg.InstallID != "" {
			installID = cfg.InstallID
		}
		if cfg.APIToken != "" {
			authed = "yes (API token present in ~/.nuon)"
		}
	}

	b.WriteString("# Nuon agent context\n\n")
	b.WriteString("Use this document to orient before creating or changing Nuon resources.\n\n")

	b.WriteString("## Current CLI selection\n\n")
	b.WriteString(fmt.Sprintf("| Field | Value |\n| --- | --- |\n"))
	b.WriteString(fmt.Sprintf("| Authenticated | %s |\n", authed))
	b.WriteString(fmt.Sprintf("| API URL | `%s` |\n", apiURL))
	b.WriteString(fmt.Sprintf("| MCP HTTP URL | `%s` |\n", mcpURL))
	b.WriteString(fmt.Sprintf("| Org ID | `%s` |\n", orgID))
	b.WriteString(fmt.Sprintf("| App ID | `%s` |\n", appID))
	b.WriteString(fmt.Sprintf("| Install ID | `%s` |\n\n", installID))

	b.WriteString("If auth or org is missing:\n\n")
	b.WriteString("```bash\nnuon auth login\nnuon orgs select\nnuon apps select\nnuon installs select   # optional\n```\n\n")

	b.WriteString("## How to call MCP\n\n")
	b.WriteString("### Preferred: local stdio proxy\n\n")
	b.WriteString("Runs on the user's machine, injects token + org ID, forwards tools to the control plane:\n\n")
	b.WriteString("```bash\nnuon agents mcp\n# writes:\nnuon agents mcp --allow-writes\n```\n\n")
	b.WriteString("MCP client config:\n\n")
	b.WriteString("```json\n{\"mcpServers\": {\"nuon\": {\"command\": \"nuon\", \"args\": [\"agents\", \"mcp\"]}}}\n```\n\n")
	b.WriteString("### Direct: control-plane HTTP MCP\n\n")
	b.WriteString(fmt.Sprintf("Point an MCP HTTP client at `%s` with:\n\n", mcpURL))
	b.WriteString("- `Authorization: Bearer <api_token>`\n")
	b.WriteString("- `X-Nuon-Org-ID: <org_id>` (recommended; required for multi-org accounts without a prior `select_org`)\n\n")
	b.WriteString("The server is **stateless** Streamable HTTP (`Stateless: true`): no durable `Mcp-Session-Id`, POST-only.\n\n")

	b.WriteString("## Creating something new (starter checklist)\n\n")
	b.WriteString("1. Confirm org/app context (`nuon agents context` / `whoami` / `list_orgs`).\n")
	b.WriteString("2. Prefer MCP tools for reads; use `--allow-writes` (or a write-scoped token) only when mutating.\n")
	b.WriteString("3. Typical create flow:\n")
	b.WriteString("   - App config in a local directory → `nuon apps sync` (or ask the user to sync).\n")
	b.WriteString("   - Install → create via dashboard/CLI/MCP write tools when available.\n")
	b.WriteString("   - Deploys / workflow steps → list workflows, pending approvals, then approve/reject/retry/cancel as needed.\n")
	b.WriteString("4. Do not invent IDs; resolve names via `list_*` / `get_*` tools first.\n")
	b.WriteString("5. Keep responses trimmed; MCP tools already return compact JSON.\n\n")

	b.WriteString("## Tool categories (control plane)\n\n")
	b.WriteString("| Domain | Examples |\n| --- | --- |\n")
	b.WriteString("| Orgs | `whoami`, `list_orgs`, `select_org` |\n")
	b.WriteString("| Apps | `list_apps`, `get_app` |\n")
	b.WriteString("| Components | `list_components`, `get_component`, `list_builds`, `get_build` |\n")
	b.WriteString("| Installs | `list_installs`, `get_install`, workflows, deploys, approvals |\n")
	b.WriteString("| Runners | `get_workflow_step_logs` |\n")
	b.WriteString("| Runbooks | `list_runbooks`, `get_runbook` |\n")
	b.WriteString("| Writes | tools whose description starts with `WRITE OPERATION:` |\n\n")

	b.WriteString("## Deprecated CLI surface\n\n")
	b.WriteString("`nuon mcp` still works but is deprecated — use `nuon agents mcp` instead.\n\n")

	b.WriteString("## Roadmap\n\n")
	b.WriteString("`nuon agents skills` will later expose scaffolding skills for creating apps, installs, and more.\n")

	return b.String()
}
