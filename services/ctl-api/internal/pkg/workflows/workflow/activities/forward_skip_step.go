package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// ForwardSkipStepRequest is the input for forwarding a skip to a step handler workflow.
type ForwardSkipStepRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

// ForwardSkipStepResponse is the output from forwarding a skip.
type ForwardSkipStepResponse struct {
	StepID string `json:"step_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardSkipStep(ctx context.Context, req ForwardSkipStepRequest) (*ForwardSkipStepResponse, error) {
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

	handle, err := handler.UpdateWithStart(ctx, a.tClient, &qs, handler.UpdateWithStartOptions{
		UpdateName:   "skip-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args:         []any{&struct{}{}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send skip-step update to step %s: %w", req.StepID, err)
	}

	var resp struct{}
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("skip-step failed on step %s: %w", req.StepID, err)
	}

	return &ForwardSkipStepResponse{StepID: req.StepID}, nil
}
