package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"whoami",
		"Who am I",
		"Get the current authenticated account and the org on this request. Returns account ID, email, an org count, and the current org when X-Nuon-Org-ID was sent. Does not list every org. Use list_orgs to find one.",
	), s.mcpWhoami)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_orgs",
		"List organizations",
		"List the organizations the authenticated account can access. Does not require X-Nuon-Org-ID. Pass query to match a name or id."+apiPkg.MCPListToolHint,
	), s.mcpListOrgs)
}
