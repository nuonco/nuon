package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_install_actions",
		"List install actions",
		"List actions available on an install (name or ID). Returns action workflow IDs and names that can be inspected with get_action or run with run_action.",
	), s.mcpListInstallActions)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_action",
		"Get action",
		"Get an install action by action_id (action_workflow_id). Pass install as name or ID. Returns whether it can be triggered manually and the install-pinned action_workflow_config_id.",
	), s.mcpGetAction)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"run_action",
		"Run action",
		"WRITE OPERATION: Trigger a configured action on an install. Requires a manual trigger on the action config. "+
			"Pass install (name or ID) and action_id (action_workflow_id). Optional action_workflow_config_id pins a specific config; otherwise the install-pinned or latest config is used. "+
			"Call list_available_roles (operation_type=trigger, principal_type=action, principal_id=action_id) before passing role.",
		false,
		false,
	), s.mcpRunAction)
}
