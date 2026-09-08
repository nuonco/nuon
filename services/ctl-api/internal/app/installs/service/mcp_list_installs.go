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

type mcpInstallListItem struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	AppID           string `json:"app_id,omitempty"`
	AppName         string `json:"app_name,omitempty"`
	SandboxStatus   string `json:"sandbox_status,omitempty"`
	ComponentStatus string `json:"component_status,omitempty"`
	RunnerStatus    string `json:"runner_status,omitempty"`
	CloudPlatform   string `json:"cloud_platform,omitempty"`
	CreatedAt       string `json:"created_at"`
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

	out := make([]mcpInstallListItem, 0, len(installs))
	for _, inst := range installs {
		item := mcpInstallListItem{
			ID:              inst.ID,
			Name:            inst.Name,
			AppID:           inst.AppID,
			AppName:         inst.App.Name,
			SandboxStatus:   string(inst.SandboxStatus),
			ComponentStatus: string(inst.CompositeComponentStatus),
			RunnerStatus:    string(inst.RunnerStatus),
			CloudPlatform:   string(inst.CloudPlatform),
			CreatedAt:       apiPkg.MCPTime(inst.CreatedAt),
		}
		out = append(out, item)
	}

	return apiPkg.MCPJSONResult(out)
}
