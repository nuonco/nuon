package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"whoami",
		"Who am I",
		"Get the current authenticated account and organization. Returns account ID, email, the orgs you can access, and the currently selected org. Use this first to confirm your auth context.",
	), s.mcpWhoami)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_orgs",
		"List organizations",
		"List the organizations the authenticated account can access. Use this to find an org ID to pass to select_org.",
	), s.mcpListOrgs)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"select_org",
		"Select organization",
		"Set the active organization for subsequent tool calls. Required before using org-scoped tools when the account belongs to more than one org. Takes an org_id from list_orgs.",
		false,
		true,
	), s.mcpSelectOrg)
}
