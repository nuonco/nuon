package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// DenyStepApprovalRequest is the input for denying a step's approval and forwarding the denial.
type DenyStepApprovalRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

// DenyStepApprovalResponse is the output from denying a step's approval.
type DenyStepApprovalResponse struct {
	StepID string `json:"step_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) DenyStepApproval(ctx context.Context, req DenyStepApprovalRequest) (*DenyStepApprovalResponse, error) {
	// Find the step's approval
	var approval app.WorkflowStepApproval
	res := a.db.WithContext(ctx).
		Where("install_workflow_step_id = ?", req.StepID).
		Preload("Response").
		Order("created_at DESC").
		First(&approval)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find approval for step %s: %w", req.StepID, res.Error)
	}

	// Only create a denial if there's no existing response
	if approval.Response == nil {
		response := app.WorkflowStepApprovalResponse{
			InstallWorkflowStepApprovalID: approval.ID,
			Type:                          app.WorkflowStepApprovalResponseTypeRetryPlan,
			Note:                          "Plan retry requested",
		}
		if res := a.db.WithContext(ctx).Create(&response); res.Error != nil {
			return nil, fmt.Errorf("unable to create denial response for step %s: %w", req.StepID, res.Error)
		}

		// Forward the denial to the step handler workflow (not the inner signal).
		// Both use the same OwnerID/OwnerType, so filter by signal type.
		var qs app.QueueSignal
		res := a.db.WithContext(ctx).
			Where(app.QueueSignal{
				OwnerID:   req.StepID,
				OwnerType: (&app.WorkflowStep{}).TableName(),
				Type:      signal.SignalType("execute-workflow-step"),
			}).
			Order("created_at DESC").
			First(&qs)
		if res.Error != nil {
			return nil, fmt.Errorf("unable to find step queue signal for step %s: %w", req.StepID, res.Error)
		}

		handle, err := a.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
			tclient.UpdateWorkflowOptions{
				WorkflowID:   qs.Workflow.ID,
				UpdateName:   "approve-plan",
				WaitForStage: tclient.WorkflowUpdateStageAccepted,
				Args: []any{
					struct {
						ApprovalResponseID string `json:"approval_response_id"`
						ResponseType       string `json:"response_type"`
					}{
						ApprovalResponseID: response.ID,
						ResponseType:       string(app.WorkflowStepApprovalResponseTypeRetryPlan),
					},
				},
			})
		if err != nil {
			return nil, fmt.Errorf("unable to send deny update to step %s: %w", req.StepID, err)
		}

		var result error
		if err := handle.Get(ctx, &result); err != nil {
			return nil, fmt.Errorf("deny update failed for step %s: %w", req.StepID, err)
		}
	}

	return &DenyStepApprovalResponse{StepID: req.StepID}, nil
}
