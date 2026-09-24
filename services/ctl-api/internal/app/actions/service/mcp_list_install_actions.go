package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListInstallActionsInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
}

type mcpInstallActionSummary struct {
	ID               string `json:"id"`
	ActionWorkflowID string `json:"action_workflow_id"`
	Name             string `json:"name"`
	CreatedAt        string `json:"created_at"`
}

func (s *service) mcpListInstallActions(ctx context.Context, _ *mcp.CallToolRequest, in mcpListInstallActionsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install: %w", err)
	}

	var actions []app.InstallActionWorkflow
	err = s.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{OrgID: orgID, InstallID: install.ID}).
		Preload("ActionWorkflow").
		Order("created_at DESC").
		Limit(50).
		Find(&actions).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list install actions: %w", err)
	}

	out := make([]mcpInstallActionSummary, 0, len(actions))
	for _, a := range actions {
		out = append(out, mcpInstallActionSummary{
			ID:               a.ID,
			ActionWorkflowID: a.ActionWorkflowID,
			Name:             a.ActionWorkflow.Name,
			CreatedAt:        apiPkg.MCPTime(a.CreatedAt),
		})
	}

	return apiPkg.MCPJSONResult(out)
}
