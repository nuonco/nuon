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
	InstallID string `json:"install_id" jsonschema:"install ID to list actions for"`
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

	var actions []app.InstallActionWorkflow
	err = s.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{OrgID: orgID, InstallID: in.InstallID}).
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
			CreatedAt:        a.CreatedAt.String(),
		})
	}

	return apiPkg.MCPJSONResult(out)
}
