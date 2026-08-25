package service

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_apps",
		Description: "List all apps in the current org with their components. Returns app name, ID, description, and associated components.",
	}, s.mcpListApps)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_app",
		Description: "Get an app by name or ID. Returns full app details including components, sandbox configs, and org info. Accepts either the app name or ID.",
	}, s.mcpGetApp)
}
