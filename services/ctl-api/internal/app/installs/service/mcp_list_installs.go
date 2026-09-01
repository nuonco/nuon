package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListInstallsInput struct {
	AppID string `json:"app_id,omitempty" jsonschema:"filter installs by app ID"`
}

func (s *service) mcpListInstalls(ctx context.Context, _ *mcp.CallToolRequest, in mcpListInstallsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var installs []app.Install
	tx := s.db.WithContext(ctx).
		Preload("App").
		Where(app.Install{OrgID: orgID}).
		Order("name ASC")

	if in.AppID != "" {
		tx = tx.Where(app.Install{AppID: in.AppID})
	}

	if res := tx.Find(&installs); res.Error != nil {
		return nil, nil, fmt.Errorf("unable to list installs: %w", res.Error)
	}

	return apiPkg.MCPJSONResult(installs)
}
