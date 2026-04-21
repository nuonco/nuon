package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// ForwardCancelStepRequest is the input for forwarding a cancellation to a step handler workflow.
type ForwardCancelStepRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

// ForwardCancelStepResponse is the output from forwarding a cancellation.
type ForwardCancelStepResponse struct {
	StepID string `json:"step_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardCancelStep(ctx context.Context, req ForwardCancelStepRequest) (*ForwardCancelStepResponse, error) {
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

	// Send the cancel update to the step's handler workflow
	handle, err := a.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   handler.CancelUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args:         []any{&handler.CancelRequest{}},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send cancel update to step %s: %w", req.StepID, err)
	}

	var result handler.CancelResponse
	if err := handle.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("cancel update failed for step %s: %w", req.StepID, err)
	}

	return &ForwardCancelStepResponse{StepID: req.StepID}, nil
}
