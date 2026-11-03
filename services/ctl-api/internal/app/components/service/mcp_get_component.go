package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetComponentInput struct {
	Component string `json:"component" jsonschema:"component name or ID"`
}

func (s *service) mcpGetComponent(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetComponentInput) (*mcp.CallToolResult, any, error) {
	component, err := require.ReadOperation(ctx, s.db, require.ReadContract{
		Permission: permissions.AppRead,
		Targets:    []require.ReadTarget{{Name: "component_id", Kind: permissions.KindComponent}},
	}, map[string]string{"component_id": in.Component}, func(ctx context.Context) (*app.Component, error) {
		return s.readComponent(ctx, "component_id")
	})
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(component)
}
