package service

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "whoami",
		Description: "Get the current authenticated account and organization. Returns account ID, email, and org ID/name. Use this first to confirm your auth context.",
	}, s.mcpWhoami)
}
