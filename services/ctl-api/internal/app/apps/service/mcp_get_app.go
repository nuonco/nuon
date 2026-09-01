package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetAppInput struct {
	App string `json:"app" jsonschema:"app name or ID"`
}

func (s *service) mcpGetApp(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetAppInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	result, err := s.findApp(ctx, orgID, in.App)
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(result)
}
