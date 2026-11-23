package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetAppAllowedHostsInput struct {
	App string `json:"app" jsonschema:"app name or ID"`
}

func (s *service) mcpGetAppAllowedHosts(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetAppAllowedHostsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	if _, err := s.findApp(ctx, orgID, in.App); err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(newAppAllowedHosts(s.cfg.URLPolicy))
}
