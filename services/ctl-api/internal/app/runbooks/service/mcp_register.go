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
}
