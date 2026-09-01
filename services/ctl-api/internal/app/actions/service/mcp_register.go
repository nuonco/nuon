package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_install_actions",
		"List install actions",
		"List actions available on an install. Returns action workflow IDs and names that can be inspected with get_action or run with run_action.",
	), s.mcpListInstallActions)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_action",
		"Get action",
		"Get an install action by action_id (action_workflow_id). Returns whether it can be triggered manually and the install-pinned action_workflow_config_id.",
	), s.mcpGetAction)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"run_action",
		"Run action",
		"WRITE OPERATION: Trigger a configured action on an install. Requires a manual trigger on the action config. "+
			"Pass install_id and action_id (action_workflow_id). Resolves the install-pinned config like the CLI create-run flow.",
		false,
		false,
	), s.mcpRunAction)
}
