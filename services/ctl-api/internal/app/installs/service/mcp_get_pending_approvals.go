package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetPendingApprovalsInput struct{}

func (s *service) mcpGetPendingApprovals(ctx context.Context, _ *mcp.CallToolRequest, _ mcpGetPendingApprovalsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var approvals []app.WorkflowStepApproval
	err = s.db.WithContext(ctx).
		Preload("InstallWorkflowStep").
		Where(app.WorkflowStepApproval{OrgID: orgID}).
		Where("id NOT IN (SELECT install_workflow_step_approval_id FROM install_workflow_step_approval_responses WHERE deleted_at = 0)").
		Order("created_at DESC").
		Find(&approvals).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list pending approvals: %w", err)
	}

	return apiPkg.MCPJSONResult(approvals)
}
