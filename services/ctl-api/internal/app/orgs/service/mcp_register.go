package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"whoami",
		"Who am I",
		"Get the current authenticated account and credential-visible organization count. Returns the current org's ID and name only when it is accessible with this credential. Does not list every org. Use list_orgs to find one.",
	), s.mcpWhoami)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_orgs",
		"List organizations",
		"List organization IDs and names visible through current grants and credential limits, including parent navigation for scoped app/install access. This does not grant org-wide read access. Does not require X-Nuon-Org-ID. Pass query to match a name or id."+apiPkg.MCPListToolHint,
	), s.mcpListOrgs)
}
