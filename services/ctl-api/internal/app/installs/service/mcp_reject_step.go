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

	target, err := s.resolveWorkflowApproval(ctx, orgID, in.ApprovalID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find approval %q: %w", in.ApprovalID, err)
	}
	approval := target.approval

	if approval.Response != nil {
		return nil, nil, fmt.Errorf("approval already has a response")
	}

	response, err := s.createWorkflowStepApprovalResponse(ctx, approval, &CreateWorkflowStepApprovalResponseRequest{
		ResponseType: app.WorkflowStepApprovalResponseTypeDeny,
		Note:         in.Reason,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create approval response: %w", err)
	}

	if err := s.notifyWorkflowOfApprovalResponse(ctx, target, response); err != nil {
		return nil, nil, err
	}

	workflowID := target.workflow.ID
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
