package service

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_workflow_step_logs",
		Description: "Get logs for a workflow step. Resolves the step's target (deploy, sandbox run, or action run) " +
			"to its log stream and returns log records from ClickHouse. Returns newest logs first by default. " +
			"Use the cursor for pagination when has_more is true.",
	}, s.mcpGetWorkflowStepLogs)
}
