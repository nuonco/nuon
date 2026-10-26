package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type mcpGetDeployInput struct {
	DeployID string `json:"deploy_id" jsonschema:"deploy ID"`
}

type mcpDeployDetail struct {
	ComponentName     string              `json:"component_name,omitempty"`
	ComponentType     string              `json:"component_type,omitempty"`
	Status            string              `json:"status"`
	StatusDescription string              `json:"status_description,omitempty"`
	Commit            *mcpDeployCommit    `json:"commit,omitempty"`
	AppBranchRun      *mcpDeployBranchRun `json:"app_branch_run,omitempty"`
	CreatedAt         string              `json:"created_at"`
	PlannedAt         string              `json:"planned_at,omitempty"`
	AppliedAt         string              `json:"applied_at,omitempty"`
	NextActions       []mcpNextAction     `json:"next_actions,omitempty"`
	ID                string              `json:"id"`
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
		Preload("ComponentBuild").
		Preload("ComponentBuild.VCSConnectionCommit").
		Preload("ComponentBuild.AppBranchRun").
		Preload("ComponentBuild.AppBranchRun.AppBranch").
		Where(app.InstallDeploy{OrgID: orgID}).
		Where("id = ?", in.DeployID).
		First(&deploy).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find deploy %q: %w", in.DeployID, err)
	}
	ctx = cctx.SetInstallIDContext(ctx, deploy.InstallID)

	summary := mcpDeploySummaryFrom(deploy.InstallID, deploy)
	detail := mcpDeployDetail{
		ComponentName:     summary.ComponentName,
		ComponentType:     summary.ComponentType,
		Status:            summary.Status,
		StatusDescription: summary.StatusDescription,
		Commit:            summary.Commit,
		AppBranchRun:      summary.AppBranchRun,
		CreatedAt:         summary.CreatedAt,
		NextActions:       summary.NextActions,
		ID:                summary.ID,
	}
	if deploy.PlannedAt != nil {
		detail.PlannedAt = apiPkg.MCPTimePtr(deploy.PlannedAt)
	}
	if deploy.AppliedAt != nil {
		detail.AppliedAt = apiPkg.MCPTimePtr(deploy.AppliedAt)
	}

	return apiPkg.MCPJSONResult(detail)
}
