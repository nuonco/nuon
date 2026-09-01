package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_workflow_step_logs",
		"Get workflow step logs",
		"Get logs for a workflow step. Resolves the step's target (deploy, sandbox run, or action run) "+
			"to its log stream and returns log records from ClickHouse. Returns newest logs first by default. "+
			"Use the cursor for pagination when has_more is true.",
	), s.mcpGetWorkflowStepLogs)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_deploy_logs",
		"Get deploy logs",
		"Get logs for an install deploy by deploy_id. Returns log records from ClickHouse, newest first. "+
			"Use the cursor for pagination when has_more is true.",
	), s.mcpGetDeployLogs)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_build_logs",
		"Get build logs",
		"Get logs for a component build by build_id. Returns log records from ClickHouse, newest first. "+
			"Use the cursor for pagination when has_more is true.",
	), s.mcpGetBuildLogs)
}
