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

Register with an MCP client:

  claude mcp add --transport stdio nuon -- nuon agents mcp
  amp mcp add nuon -- nuon agents mcp

Cursor / Cursor Agent: write ~/.cursor/mcp.json then agent mcp enable nuon.

  claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes
  amp mcp add nuon -- nuon agents mcp --allow-writes

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

  claude mcp add --transport stdio nuon -- nuon agents mcp
  amp mcp add nuon -- nuon agents mcp
  # Cursor: write ~/.cursor/mcp.json, then agent mcp enable nuon
  # writes:
  claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes
  amp mcp add nuon -- nuon agents mcp --allow-writes

Or add to .mcp.json:

  {"mcpServers": {"nuon": {"command": "nuon", "args": ["agents", "mcp"]}}}

Example — override the upstream server:

  nuon agents mcp --url https://mcp.example.com/mcp --name nuon-example

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
	cmd.AddCommand(c.mcpSetupCmd())

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
		derived, err := mcpserver.EndpointFromAPIURL(cfg.APIURL)
		if err != nil {
			mcpURL = err.Error()
		} else {
			mcpURL = derived
		}
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
	b.WriteString("Register with an MCP client:\n\n")
	b.WriteString("```bash\n")
	b.WriteString("claude mcp add --transport stdio nuon -- nuon agents mcp\n")
	b.WriteString("amp mcp add nuon -- nuon agents mcp\n")
	b.WriteString("# Cursor: write ~/.cursor/mcp.json, then:\n")
	b.WriteString("agent mcp enable nuon\n")
	b.WriteString("# writes:\n")
	b.WriteString("claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes\n")
	b.WriteString("amp mcp add nuon -- nuon agents mcp --allow-writes\n")
	b.WriteString("```\n\n")
	b.WriteString("Cursor JSON (`~/.cursor/mcp.json` or `.cursor/mcp.json`):\n\n")
	b.WriteString("```json\n{\"mcpServers\": {\"nuon\": {\"command\": \"nuon\", \"args\": [\"agents\", \"mcp\"]}}}\n```\n\n")
	b.WriteString(fmt.Sprintf("The upstream URL above (`%s`) comes from `api_url` in the CLI config. Override it with `--url`, and the name in the client's MCP list with `--name`:\n\n", mcpURL))
	b.WriteString("```bash\nnuon agents mcp --url https://mcp.example.com/mcp --name nuon-example\n```\n\n")
	b.WriteString("`--url`, `--name`, and `--allow-writes` pass through `nuon agents mcp setup`, which also copies a non-default `-C` config path into the command it writes.\n\n")
	b.WriteString("The proxy sets `X-Nuon-Org-ID`. Do not call `select_org` unless the header is missing.\n\n")
	b.WriteString("### Direct: control-plane HTTP MCP\n\n")
	b.WriteString(fmt.Sprintf("Point an MCP HTTP client at `%s` with:\n\n", mcpURL))
	b.WriteString("- `Authorization: Bearer <api_token>`\n")
	b.WriteString("- `X-Nuon-Org-ID: <org_id>` (required for multi-org accounts without `select_org`)\n\n")
	b.WriteString("The server is **stateless** Streamable HTTP: no durable `Mcp-Session-Id`, POST-only.\n\n")

	b.WriteString("## Creating something new (starter checklist)\n\n")
	b.WriteString("1. Confirm org/app context (`nuon agents context` / `whoami`).\n")
	b.WriteString("2. Prefer MCP tools for reads; use `--allow-writes` (or a write-scoped token) only when mutating.\n")
	b.WriteString("3. Typical create flow:\n")
	b.WriteString("   - App config in a local directory → `nuon apps sync` (or ask the user to sync).\n")
	b.WriteString("   - Installs: dashboard or CLI (`nuon installs create`).\n")
	b.WriteString("   - Deploys / workflow steps → list workflows and pending approvals, then approve/reject/retry/cancel as needed.\n")
	b.WriteString("4. Do not invent IDs; resolve names via `list_*` / `get_*` tools first.\n")
	b.WriteString("5. Keep responses trimmed; MCP tools already return compact JSON.\n\n")

	b.WriteString("## Tools (control plane)\n\n")
	b.WriteString("Writes are hidden from the stdio proxy unless `--allow-writes` is set. Descriptions start with `WRITE OPERATION:`.\n\n")
	b.WriteString("| Domain | Read | Write |\n| --- | --- | --- |\n")
	b.WriteString("| Orgs | `whoami`, `list_orgs`, `select_org` | |\n")
	b.WriteString("| Apps | `list_apps`, `get_app`, `list_app_branches`, `get_app_branch`, `list_app_branch_preview_sources` | `preview_app_branch` |\n")
	b.WriteString("| Components | `list_components`, `get_component`, `list_builds`, `get_build` | |\n")
	b.WriteString("| Installs | `list_installs`, `get_install`, `list_install_components`, `get_install_inputs`, `list_workflows`, `get_workflow`, `get_workflow_step`, `watch_workflow`, `get_pending_approvals`, `list_deploys`, `get_deploy` | `update_install_inputs`, `deploy_install_components`, `reprovision_install`, `reprovision_sandbox`, `deprovision_install`, `deprovision_sandbox`, `approve_step`, `reject_step`, `retry_step`, `cancel_workflow` |\n")
	b.WriteString("| Actions | `list_install_actions`, `get_action` | `run_action` |\n")
	b.WriteString("| Logs | `get_workflow_step_logs`, `get_deploy_logs`, `get_build_logs` | |\n")
	b.WriteString("| Runbooks | `list_runbooks`, `get_runbook` | |\n\n")
	b.WriteString("Catalog: https://docs.nuon.co/guides/agents/tools\n\n")

	b.WriteString("## Sample queries\n\n")
	b.WriteString("Paste these into the LLM client after MCP is connected. The client should pick tools; do not require the user to name them. Writes need `--allow-writes`.\n\n")
	b.WriteString("- What's in this org? What apps and installs are there, and what's each install's status?\n")
	b.WriteString("- What's the status of install <name>? Include components, recent workflows, and anything unhealthy.\n")
	b.WriteString("- Are there any pending workflow approvals? For each, tell me the install, workflow, and what is waiting.\n")
	b.WriteString("- Why did the latest deploy on install <name> fail? Quote the relevant log lines.\n")
	b.WriteString("- Why did the latest build for component <name> fail?\n")
	b.WriteString("- Watch workflow <id> until it finishes and tell me when the status changes.\n")
	b.WriteString("- What actions can I run on install <name>?\n")
	b.WriteString("- Show the runbooks for app <name>.\n")
	b.WriteString("- What's the latest on the default branch for app <name>?\n")
	b.WriteString("- Preview PR <number> for app <name> against install <name> (plan-only unless I ask to apply).\n")
	b.WriteString("- Review the pending plan on workflow <id>, then ask me before approving or rejecting (write).\n")
	b.WriteString("- Plan a reprovision of install <name>. Do not apply until I confirm (write).\n")
	b.WriteString("- What are the current inputs on install <name>? Propose an update and wait for approval before writing (write).\n\n")
	b.WriteString("More: https://docs.nuon.co/guides/agents/sample-queries\n\n")

	b.WriteString("## Deprecated CLI surface\n\n")
	b.WriteString("`nuon mcp` is deprecated — use `nuon agents mcp`. It still proxies when an MCP client drives it over piped stdio, but exits immediately when run from a terminal.\n")

	return b.String()
}
