package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// ForwardIsRetryableRequest is the input for forwarding an is-retryable query to a step handler workflow.
type ForwardIsRetryableRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

// ForwardIsRetryableResponse is the output from the is-retryable update.
type ForwardIsRetryableResponse struct {
	Retryable  bool   `json:"retryable"`
	Skippable  bool   `json:"skippable"`
	AutoRetry  bool   `json:"auto_retry"`
	MaxRetries int    `json:"max_retries"`
	RetryGroup bool   `json:"retry_group"`
	RetryIndex int    `json:"retry_index"`
	StepID     string `json:"step_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardIsRetryable(ctx context.Context, req ForwardIsRetryableRequest) (*ForwardIsRetryableResponse, error) {
	// Find the step's handler workflow via the queue_signals table.
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

	// Send the is-retryable update via update-with-start.
	rawResp, err := handler.UpdateWithStart(ctx, a.tClient, &qs, handler.UpdateWithStartOptions{
		UpdateName:   "is-retryable",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send is-retryable update to step %s: %w", req.StepID, err)
	}

	var resp ForwardIsRetryableResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("is-retryable update failed for step %s: %w", req.StepID, err)
	}

	return &resp, nil
}
