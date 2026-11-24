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

	target, err := s.resolveWorkflowApproval(ctx, orgID, in.ApprovalID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find approval %q: %w", in.ApprovalID, err)
	}
	approval := target.approval

	if approval.Response != nil {
		return nil, nil, fmt.Errorf("approval already has a response")
	}

	response, err := s.createWorkflowStepApprovalResponse(ctx, approval, &CreateWorkflowStepApprovalResponseRequest{
		ResponseType: app.WorkflowStepApprovalResponseTypeApprove,
		Note:         in.Note,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create approval response: %w", err)
	}

	if err := s.notifyWorkflowOfApprovalResponse(ctx, target, response); err != nil {
		return nil, nil, err
	}

	workflowID := target.workflow.ID
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

func (s *service) notifyWorkflowOfApprovalResponse(ctx context.Context, target *workflowApprovalTarget, response *app.WorkflowStepApprovalResponse) error {
	if err := s.dispatchApprovalResponseSignal(ctx, target, response); err != nil {
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
