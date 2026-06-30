package service

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_runbooks",
		Description: "List all runbooks for an app. Runbooks define ordered sequences of deploy and action steps that can be executed on installs.",
	}, s.mcpListRunbooks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_runbook",
		Description: "Get a runbook by ID with its configuration and steps. Returns the runbook definition including the ordered sequence of operations.",
	}, s.mcpGetRunbook)
}
