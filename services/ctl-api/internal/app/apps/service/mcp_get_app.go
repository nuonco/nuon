package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type mcpGetAppInput struct {
	App string `json:"app" jsonschema:"app name or ID"`
}

func (s *service) mcpGetApp(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetAppInput) (*mcp.CallToolResult, any, error) {
	result, err := s.getApp(ctx, in.App)
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(result)
}
