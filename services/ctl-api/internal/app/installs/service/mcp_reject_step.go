package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpRejectStepInput struct {
	ApprovalID string `json:"approval_id" jsonschema:"the approval ID to reject"`
	Reason     string `json:"reason,omitempty" jsonschema:"reason for rejection"`
}

func (s *service) mcpRejectStep(ctx context.Context, _ *mcp.CallToolRequest, in mcpRejectStepInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.ApprovalID == "" {
		return nil, nil, fmt.Errorf("approval_id is required")
	}

	var approval app.WorkflowStepApproval
	err = s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", in.ApprovalID, orgID).
		Preload("InstallWorkflowStep").
		Preload("Response").
		First(&approval).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find approval %q: %w", in.ApprovalID, err)
	}

	if approval.Response != nil {
		return nil, nil, fmt.Errorf("approval already has a response")
	}

	response := app.WorkflowStepApprovalResponse{
		InstallWorkflowStepApprovalID: approval.ID,
		Type:                          app.WorkflowStepApprovalResponseTypeDeny,
		Note:                          in.Reason,
	}
	if err := s.db.WithContext(ctx).Create(&response).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to create approval response: %w", err)
	}

	if err := s.notifyWorkflowOfApprovalResponse(ctx, approval, &response); err != nil {
		return nil, nil, err
	}

	workflowID := approval.InstallWorkflowStep.InstallWorkflowID
	out := map[string]any{
		"status":      "rejected",
		"approval_id": approval.ID,
		"response_id": response.ID,
		"workflow_id": workflowID,
	}
	if next := mcpWatchContinuation(workflowID); next != nil {
		out["next_action"] = next
	}
	return apiPkg.MCPJSONResult(out)
}
