package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListAppsInput struct{}

func (s *service) mcpListApps(ctx context.Context, _ *mcp.CallToolRequest, _ mcpListAppsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var apps []*app.App
	org := &app.Org{ID: orgID}

	err = s.db.WithContext(ctx).
		Preload("Components").
		Order("apps.name ASC").
		Model(org).Association("Apps").Find(&apps)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list apps: %w", err)
	}

	return apiPkg.MCPJSONResult(apps)
}
