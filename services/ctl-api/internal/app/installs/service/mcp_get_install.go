package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetInstallInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
}

func (s *service) mcpGetInstall(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetInstallInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(install)
}
