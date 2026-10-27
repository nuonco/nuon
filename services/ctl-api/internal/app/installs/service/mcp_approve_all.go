package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpApproveAllInput struct {
	WorkflowID string `json:"workflow_id" jsonschema:"workflow ID to switch to approve-all"`
}

func (s *service) mcpApproveAll(ctx context.Context, _ *mcp.CallToolRequest, in mcpApproveAllInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.WorkflowID == "" {
		return nil, nil, fmt.Errorf("workflow_id is required")
	}

	var workflow app.Workflow
	err = s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", in.WorkflowID, orgID).
		Preload("Steps").
		First(&workflow).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find workflow %q: %w", in.WorkflowID, err)
	}
	if err := mcpApproveAllGate(workflow); err != nil {
		return nil, nil, err
	}

	res := s.db.WithContext(ctx).
		Model(&app.Workflow{}).
		Where("id = ? AND org_id = ?", workflow.ID, orgID).
		Update("approval_option", app.InstallApprovalOptionApproveAll)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to set approve-all: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, nil, fmt.Errorf("unable to set approve-all on workflow %q", workflow.ID)
	}

	s.markWorkflowApproveAll(ctx, workflow.ID)
	s.l.Info("workflow approve-all requested",
		zap.String("flow_event", "workflow.approve_all_requested"),
		zap.String("workflow_id", workflow.ID),
	)

	return apiPkg.MCPJSONResult(map[string]string{
		"workflow_id":     workflow.ID,
		"approval_option": string(app.InstallApprovalOptionApproveAll),
		"status":          "approve-all",
	})
}

func mcpApproveAllGate(workflow app.Workflow) error {
	if workflow.ApprovalOption == app.InstallApprovalOptionApproveAll {
		return fmt.Errorf("workflow approval_option is already approve-all")
	}
	if workflow.PlanOnly {
		return fmt.Errorf("approve-all is not available for a drift scan")
	}
	if workflow.Finished {
		return fmt.Errorf("workflow is already finished")
	}
	if workflow.Status.Status == app.StatusCancelled {
		return fmt.Errorf("workflow is cancelled")
	}
	if workflow.ApprovalOption != app.InstallApprovalOptionPrompt {
		return fmt.Errorf("workflow approval_option is %q", workflow.ApprovalOption)
	}
	if !mcpWorkflowAllowsApproveAll(workflow) {
		return fmt.Errorf("workflow has no approval steps")
	}
	return nil
}
