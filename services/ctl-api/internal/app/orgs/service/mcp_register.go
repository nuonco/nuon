package service

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "whoami",
		Description: "Get the current authenticated account and organization. Returns account ID, email, the orgs you can access, and the currently selected org. Use this first to confirm your auth context.",
	}, s.mcpWhoami)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_orgs",
		Description: "List the organizations the authenticated account can access. Use this to find an org ID to pass to select_org.",
	}, s.mcpListOrgs)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "select_org",
		Description: "Set the active organization for this session. Required before using org-scoped tools when the account belongs to more than one org. Takes an org_id from list_orgs.",
	}, s.mcpSelectOrg)
}
