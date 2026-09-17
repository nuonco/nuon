package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type service struct{}

var _ api.MCPService = (*service)(nil)

func New() *service {
	return &service{}
}

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, api.MCPReadTool(
		"whoami",
		"Who am I",
		"Get the authenticated Nuon employee account and organization context. Use this first to confirm nuonctl MCP authentication.",
	), s.mcpWhoami)
}
