package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/access"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetComponentInput struct {
	Component string `json:"component" jsonschema:"component name or ID"`
}

func (s *service) mcpGetComponent(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetComponentInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.MCPRead(ctx)
	if err != nil {
		return nil, nil, err
	}

	component, err := s.findComponent(ctx, orgID, in.Component)
	if err != nil {
		return nil, nil, err
	}
	if _, err := require.MCPPoint(ctx, s.db, access.Components, permissions.ComponentInspect, component.ID); err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(component)
}
