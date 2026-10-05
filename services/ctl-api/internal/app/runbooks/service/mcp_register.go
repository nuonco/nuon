package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_runbooks",
		"List runbooks",
		"List runbooks. Pass install (name or ID) to list runbooks pinned to that install's current app config, including runbook_config_id and step_count. Pass app_id to list the app catalog. synced defaults to true for an install; false lists runbooks no longer in the install's app config. Use get_runbook before run_runbook."+apiPkg.MCPListToolHint,
	), s.mcpListRunbooks)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_runbook",
		"Get runbook",
		"Get a runbook with the ordered inputs and steps needed to run it. Pass install (name or ID) plus runbook (name or ID) to load the config pinned to that install. Each step includes step_id, name, and type. runnable is false when the runbook is not in the install's app config. Without install, step IDs come from the latest app config and may not match an install.",
	), s.mcpGetRunbook)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"run_runbook",
		"Run runbook",
		"WRITE OPERATION: Run a configured runbook on an install. This starts an install workflow of type runbook_run. "+
			"Do not call this tool until the user has chosen which steps to run. First call get_runbook with the install and show every step as a numbered checklist (name and type), all enabled. Let the user turn steps off. At least one must stay enabled. Show inputs and collect required values. "+
			"Execution role is optional: call list_available_roles (operation_type=trigger) and offer the returned names, or omit role for the default. A role configured on a step still applies. "+
			"After the user confirms, call this tool. Omit steps only when every step stays enabled. To skip any, pass every step_id from get_runbook and set enabled false for the ones they turned off; a step left out of steps still runs. "+
			"Returns run_id, workflow_id, and next_action. Ask the user whether they want to watch the workflow. Call next_action only if they choose to watch. Once watching, each watch returns when a step starts or finishes. After every watch result, the next user-visible message is step_progress, then call watch again with the returned cursor. A row of watch calls with no step_progress between them is wrong.",
		true,
		false,
	), s.mcpRunRunbook)
}
