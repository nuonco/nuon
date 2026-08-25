package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpGetComponentInput struct {
	Component string `json:"component" jsonschema:"component name or ID"`
}

func (s *service) mcpGetComponent(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetComponentInput) (*mcp.CallToolResult, any, error) {
	orgID := keys.OrgIDFromContext(ctx)

	component, err := s.findComponent(ctx, orgID, in.Component)
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(component)
}
