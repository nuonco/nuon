package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetComponentInput struct {
	Component string `json:"component" jsonschema:"component name or ID"`
}

func (s *service) mcpGetComponent(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetComponentInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	component, err := s.findComponent(ctx, orgID, in.Component)
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(component)
}
