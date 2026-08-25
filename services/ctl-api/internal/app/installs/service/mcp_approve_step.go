package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type mcpApproveStepInput struct {
	ApprovalID string `json:"approval_id" jsonschema:"the approval ID to approve"`
}

func (s *service) mcpApproveStep(ctx context.Context, _ *mcp.CallToolRequest, in mcpApproveStepInput) (*mcp.CallToolResult, any, error) {
	if err := requireWriteScope(ctx); err != nil {
		return nil, nil, err
	}

	orgID := keys.OrgIDFromContext(ctx)

	var approval app.WorkflowStepApproval
	err := s.db.WithContext(ctx).
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
	}
	if err := s.db.WithContext(ctx).Create(&response).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to create approval response: %w", err)
	}

	stepID := approval.InstallWorkflowStepID
	var step app.WorkflowStep
	if err := s.db.WithContext(ctx).First(&step, "id = ?", stepID).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find step: %w", err)
	}

	if err := s.flowsClient.ApprovePlan(ctx, &flowclient.ApprovePlanRequest{
		InstallWorkflowID:  step.OwnerID,
		StepID:             stepID,
		ApprovalResponseID: response.ID,
		ResponseType:       app.WorkflowStepApprovalResponseTypeApprove,
	}); err != nil {
		s.l.Warn("failed to dispatch approval", zap.Error(err))
	}

	return apiPkg.MCPJSONResult(map[string]string{
		"status":      "approved",
		"approval_id": approval.ID,
		"response_id": response.ID,
	})
}
