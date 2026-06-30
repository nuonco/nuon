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
}
