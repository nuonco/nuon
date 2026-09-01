package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_components",
		"List components",
		"List all components in the org. Optionally filter by app_id. Returns component name, ID, type, and app association.",
	), s.mcpListComponents)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_component",
		"Get component",
		"Get a component by name or ID. Returns full component details including configs, dependencies, and app info. Accepts either the component name or ID.",
	), s.mcpGetComponent)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_builds",
		"List builds",
		"List the 20 most recent builds for a component. Returns build ID, status, git ref, and timestamp.",
	), s.mcpListBuilds)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_build",
		"Get build",
		"Get build details by ID including status, git ref, source image ref, and log stream ID for log fetching.",
	), s.mcpGetBuild)
}
