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

type mcpApproveStepInput struct {
	ApprovalID string `json:"approval_id" jsonschema:"the approval ID to approve"`
	Note       string `json:"note,omitempty" jsonschema:"optional note stored on the approval response"`
}

func (s *service) mcpApproveStep(ctx context.Context, _ *mcp.CallToolRequest, in mcpApproveStepInput) (*mcp.CallToolResult, any, error) {
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
		Type:                          app.WorkflowStepApprovalResponseTypeApprove,
		Note:                          in.Note,
	}
	if err := s.db.WithContext(ctx).Create(&response).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to create approval response: %w", err)
	}

	if err := s.notifyWorkflowOfApprovalResponse(ctx, approval, &response); err != nil {
		return nil, nil, err
	}

	workflowID := approval.InstallWorkflowStep.InstallWorkflowID
	out := map[string]any{
		"status":      "approved",
		"approval_id": approval.ID,
		"response_id": response.ID,
		"workflow_id": workflowID,
	}
	if next := mcpWatchContinuation(workflowID); next != nil {
		out["next_action"] = next
	}
	return apiPkg.MCPJSONResult(out)
}

func (s *service) notifyWorkflowOfApprovalResponse(ctx context.Context, approval app.WorkflowStepApproval, response *app.WorkflowStepApprovalResponse) error {
	workflowID := approval.InstallWorkflowStep.InstallWorkflowID
	stepID := approval.InstallWorkflowStepID
	if workflowID == "" && stepID != "" {
		var step app.WorkflowStep
		if err := s.db.WithContext(ctx).Select("install_workflow_id").First(&step, "id = ?", stepID).Error; err != nil {
			return s.abandonApprovalResponse(ctx, response, fmt.Errorf("unable to find step: %w", err))
		}
		workflowID = step.InstallWorkflowID
	}
	if workflowID == "" {
		return s.abandonApprovalResponse(ctx, response, fmt.Errorf("step %s has no workflow", stepID))
	}

	if err := s.dispatchApprovalResponseSignal(ctx, workflowID, stepID, approval.ID, response.ID, response.Type); err != nil {
		return s.abandonApprovalResponse(ctx, response, err)
	}
	return nil
}

func (s *service) abandonApprovalResponse(ctx context.Context, response *app.WorkflowStepApprovalResponse, cause error) error {
	if delErr := s.db.WithContext(ctx).Delete(response).Error; delErr != nil {
		s.l.Warn("failed to roll back approval response after dispatch failure",
			zap.String("response_id", response.ID),
			zap.Error(delErr),
		)
		return fmt.Errorf("approval response %s was saved but the workflow was not notified: %w", response.ID, cause)
	}
	return fmt.Errorf("unable to notify the workflow; the approval is still pending: %w", cause)
}
