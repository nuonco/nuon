package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_runbooks",
		"List runbooks",
		"List all runbooks for an app. Runbooks define ordered sequences of deploy and action steps that can be executed on installs.",
	), s.mcpListRunbooks)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_runbook",
		"Get runbook",
		"Get a runbook by ID with its configuration and steps. Returns the runbook definition including the ordered sequence of operations.",
	), s.mcpGetRunbook)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"run_runbook",
		"Run runbook",
		"WRITE OPERATION: Run a configured runbook on an install. Accepts install and runbook names or IDs, optional inputs, step selections, and role. Ask the user before triggering. Returns run_id and workflow_id; use watch_workflow to follow progress.",
		true,
		false,
	), s.mcpRunRunbook)
}
