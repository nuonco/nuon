package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetDeployInput struct {
	DeployID string `json:"deploy_id" jsonschema:"deploy ID"`
}

type mcpDeployDetail struct {
	ID                string `json:"id"`
	ComponentID       string `json:"component_id,omitempty"`
	ComponentName     string `json:"component_name,omitempty"`
	BuildID           string `json:"build_id"`
	Status            string `json:"status"`
	StatusDescription string `json:"status_description,omitempty"`
	LogStreamID       string `json:"log_stream_id,omitempty"`
	CreatedAt         string `json:"created_at"`
	PlannedAt         string `json:"planned_at,omitempty"`
	AppliedAt         string `json:"applied_at,omitempty"`
}

func (s *service) mcpGetDeploy(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetDeployInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var deploy app.InstallDeploy
	err = s.db.WithContext(ctx).
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Preload("LogStream").
		Where(app.InstallDeploy{OrgID: orgID}).
		Where("id = ?", in.DeployID).
		First(&deploy).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find deploy %q: %w", in.DeployID, err)
	}

	detail := mcpDeployDetail{
		ID:                deploy.ID,
		BuildID:           deploy.ComponentBuildID,
		Status:            string(deploy.Status),
		StatusDescription: deploy.StatusDescription,
		CreatedAt:         deploy.CreatedAt.String(),
	}

	if deploy.InstallComponent.Component.ID != "" {
		detail.ComponentID = deploy.InstallComponent.Component.ID
		detail.ComponentName = deploy.InstallComponent.Component.Name
	}
	if deploy.LogStream.ID != "" {
		detail.LogStreamID = deploy.LogStream.ID
	}
	if deploy.PlannedAt != nil {
		detail.PlannedAt = deploy.PlannedAt.String()
	}
	if deploy.AppliedAt != nil {
		detail.AppliedAt = deploy.AppliedAt.String()
	}

	return apiPkg.MCPJSONResult(detail)
}
