package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_apps",
		"List apps",
		"List apps in the current org with their components. Returns app name, ID, description, and associated components."+apiPkg.MCPListToolHint,
	), s.mcpListApps)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_app",
		"Get app",
		"Get an app by name or ID. Returns full app details including components, sandbox configs, and org info. Accepts either the app name or ID.",
	), s.mcpGetApp)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_app_config_schema",
		"Get app config schema",
		"Return the JSON schema for one app-config file type, the same schema the Nuon LSP uses. "+
			"Pass type such as helm, action, or runbook. The response includes the required header comment and a suggested path. "+
			"Omit type to list valid types. Use this schema when creating or editing an app config file. When you then validate the directory for this MCP server, use the Nuon CLI binary and -C value from the server instructions. This does not sync.",
	), s.mcpGetAppConfigSchema)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_app_branches",
		"List app branches",
		"List app branches for an app (name or ID). Returns each branch name, ID, and a summary of the latest run (status, whether it succeeded)."+apiPkg.MCPListToolHint,
	), s.mcpListAppBranches)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_app_branch",
		"Get app branch overview",
		"Get an overview of an app branch. Answers: did the last run succeed, what changed (config sections / git files), and how far install-group deploys have gotten (per-install status). Pass app and branch by name or ID.",
	), s.mcpGetAppBranch)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_app_branch_runs",
		"List app branch runs",
		"List recent runs for an app branch. Returns run IDs, pull request numbers, preview status, workflow IDs, and outcomes. Use this to find a specific historical or preview run."+apiPkg.MCPListToolHint,
	), s.mcpListAppBranchRuns)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_app_branch_run",
		"Get app branch run",
		"Get one app branch run with its change summary, workflow progress, and per-install deployment status. Select by run_id or pr_number; omit both for the latest run. Use the returned workflow_id with watch_workflow.",
	), s.mcpGetAppBranchRun)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_app_branch_preview_sources",
		"List app branch preview sources",
		"List preview sources for an app branch: open pull requests targeting the branch and other git branches in the repo. Use this to pick a pr_number or git_ref for preview_app_branch.",
	), s.mcpListAppBranchPreviewSources)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"preview_app_branch",
		"Preview app branch",
		"WRITE OPERATION: Trigger an app-branch preview run (same as `nuon branches preview`). "+
			"Pass pr_number (preview this PR against an install), git_ref, or app_config_id for a local synced config. "+
			"Optional config_id selects a specific app branch config (defaults to latest). "+
			"HTTP MCP cannot read the local workspace — sync with the CLI first, then pass app_config_id. "+
			"Default mode is plan-only; ask before mode=apply. Returns run_id, workflow_id, and next_action. Ask the user whether they want to watch the workflow; call next_action only if they choose to watch. Once watching, each watch returns when a step starts or finishes. After every watch result, the next user-visible message is step_progress, then call watch again with the returned cursor. A row of watch calls with no step_progress between them is wrong. Use get_app_branch_run for that exact run.",
		true,
		false,
	), s.mcpPreviewAppBranch)
}
