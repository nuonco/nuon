package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_apps",
		"List apps",
		"List all apps in the current org with their components. Returns app name, ID, description, and associated components.",
	), s.mcpListApps)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_app",
		"Get app",
		"Get an app by name or ID. Returns full app details including components, sandbox configs, and org info. Accepts either the app name or ID.",
	), s.mcpGetApp)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_app_branches",
		"List app branches",
		"List app branches for an app (name or ID). Returns each branch name, ID, and a summary of the latest run (status, whether it succeeded).",
	), s.mcpListAppBranches)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_app_branch",
		"Get app branch overview",
		"Get an overview of an app branch. Answers: did the last run succeed, what changed (config sections / git files), and how far install-group deploys have gotten (per-install status). Pass app and branch by name or ID. Use after preview_app_branch to follow preview progress.",
	), s.mcpGetAppBranch)

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
			"HTTP MCP cannot read the local workspace — sync with the CLI first, then pass app_config_id. "+
			"Default mode is plan-only; ask before mode=apply. Returns run_id and workflow_id; follow with watch_workflow and get_app_branch.",
		true,
		false,
	), s.mcpPreviewAppBranch)
}
