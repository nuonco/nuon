package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpListDeploysInput struct {
	InstallID string `json:"install_id" jsonschema:"install ID to list deploys for"`
}

type mcpDeploySummary struct {
	ID            string `json:"id"`
	ComponentID   string `json:"component_id,omitempty"`
	ComponentName string `json:"component_name,omitempty"`
	BuildID       string `json:"build_id,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

func (s *service) mcpListDeploys(ctx context.Context, _ *mcp.CallToolRequest, in mcpListDeploysInput) (*mcp.CallToolResult, any, error) {
	orgID := keys.OrgIDFromContext(ctx)

	var deploys []app.InstallDeploy
	err := s.db.WithContext(ctx).
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Where(app.InstallDeploy{OrgID: orgID}).
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where("install_components.install_id = ?", in.InstallID).
		Order("install_deploys.created_at DESC").
		Limit(20).
		Find(&deploys).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list deploys: %w", err)
	}

	out := make([]mcpDeploySummary, 0, len(deploys))
	for _, d := range deploys {
		summary := mcpDeploySummary{
			ID:        d.ID,
			BuildID:   d.ComponentBuildID,
			Status:    string(d.Status),
			CreatedAt: d.CreatedAt.String(),
		}
		if d.InstallComponent.Component.ID != "" {
			summary.ComponentID = d.InstallComponent.Component.ID
			summary.ComponentName = d.InstallComponent.Component.Name
		}
		out = append(out, summary)
	}

	return apiPkg.MCPJSONResult(out)
}
