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
		"WRITE OPERATION: Trigger a configured action on an install. This starts an install workflow of type action_workflow_run. Requires a manual trigger on the action config. "+
			"Pass install (name or ID) and action_id (action_workflow_id). Optional action_workflow_config_id pins a specific config; otherwise the install-pinned or latest config is used. "+
			"Ask the user before triggering. Call list_available_roles (operation_type=trigger, principal_type=action, principal_id=action_id) before passing role. "+
			"Returns workflow_id and next_action. Ask the user whether they want to watch the workflow. Call next_action only if they choose to watch. Once watching, each watch returns when a step starts or finishes. After every watch result, the next user-visible message is step_progress, then call watch again with the returned cursor. A row of watch calls with no step_progress between them is wrong.",
		false,
		false,
	), s.mcpRunAction)
}
