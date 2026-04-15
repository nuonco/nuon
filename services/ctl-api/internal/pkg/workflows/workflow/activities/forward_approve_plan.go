package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ForwardApprovePlanRequest is the input for forwarding an approval to a step handler workflow.
type ForwardApprovePlanRequest struct {
	StepID             string `json:"step_id" validate:"required"`
	ApprovalResponseID string `json:"approval_response_id"`
	ResponseType       string `json:"response_type"`
}

// ForwardApprovePlanResponse is the output from forwarding an approval.
type ForwardApprovePlanResponse struct {
	StepID string `json:"step_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardApprovePlan(ctx context.Context, req ForwardApprovePlanRequest) (*ForwardApprovePlanResponse, error) {
	// Find the step's handler workflow via the queue_signals table.
	// Filter by signal type to avoid matching the inner signal (same OwnerID/OwnerType).
	var qs app.QueueSignal
	res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   req.StepID,
			OwnerType: (&app.WorkflowStep{}).TableName(),
			Type:      "execute-workflow-step",
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find step queue signal for step %s: %w", req.StepID, res.Error)
	}

	// The approve-plan update arg must match executeworkflowstep.ApprovePlanRequest JSON shape.
	type approvePlanArg struct {
		ApprovalResponseID string `json:"approval_response_id"`
		ResponseType       string `json:"response_type"`
	}

	// Send the approve-plan update to the step's handler workflow
	handle, err := a.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "approve-plan",
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args: []any{
				approvePlanArg{
					ApprovalResponseID: req.ApprovalResponseID,
					ResponseType:       req.ResponseType,
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send approve-plan update to step %s: %w", req.StepID, err)
	}

	var result error
	if err := handle.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("approve-plan update failed for step %s: %w", req.StepID, err)
	}

	return &ForwardApprovePlanResponse{StepID: req.StepID}, nil
}
