package service

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_components",
		Description: "List all components in the org. Optionally filter by app_id. Returns component name, ID, type, and app association.",
	}, s.mcpListComponents)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_component",
		Description: "Get a component by name or ID. Returns full component details including configs, dependencies, and app info. Accepts either the component name or ID.",
	}, s.mcpGetComponent)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_builds",
		Description: "List the 20 most recent builds for a component. Returns build ID, status, git ref, and timestamp.",
	}, s.mcpListBuilds)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_build",
		Description: "Get build details by ID including status, git ref, source image ref, and log stream ID for log fetching.",
	}, s.mcpGetBuild)
}
