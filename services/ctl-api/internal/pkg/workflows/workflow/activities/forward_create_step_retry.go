package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ForwardCreateStepRetryRequest is the input for forwarding a create-step-retry to a step handler workflow.
type ForwardCreateStepRetryRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

// ForwardCreateStepRetryResponse is the output from forwarding a create-step-retry.
type ForwardCreateStepRetryResponse struct {
	StepID    string `json:"step_id"`
	NewStepID string `json:"new_step_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardCreateStepRetry(ctx context.Context, req ForwardCreateStepRetryRequest) (*ForwardCreateStepRetryResponse, error) {
	// Find the step's handler workflow via the queue_signals table
	var qs app.QueueSignal
	res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   req.StepID,
			OwnerType: (&app.WorkflowStep{}).TableName(),
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find step queue signal for step %s: %w", req.StepID, res.Error)
	}

	// Send the create-step-retry update to the step's handler workflow
	handle, err := a.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "create-step-retry",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send create-step-retry update to step %s: %w", req.StepID, err)
	}

	// Response shape matches executeworkflowstep.CreateStepRetryResponse
	type stepRetryResult struct {
		NewStepID string `json:"new_step_id"`
	}
	var result stepRetryResult
	if err := handle.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("create-step-retry update failed for step %s: %w", req.StepID, err)
	}

	return &ForwardCreateStepRetryResponse{StepID: req.StepID, NewStepID: result.NewStepID}, nil
}
