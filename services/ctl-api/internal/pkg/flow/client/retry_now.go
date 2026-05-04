package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstep"
)

// RetryNowRequest is the input for short-circuiting the auto-retry backoff
// on a workflow step that is currently waiting to retry.
type RetryNowRequest struct {
	StepID string
}

// RetryNowResponse is the response from the retry-now update.
type RetryNowResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// RetryNow sends a "retry-now" update to the executeworkflowstep handler
// workflow for the given step. The handler observes the resulting flag and
// breaks out of its backoff wait so the inner signal runs immediately.
//
// Should only be called when the step is in StatusWaitingToRetry — calling
// it on a step that isn't waiting is a no-op (the flag is only checked by
// the wait gate).
func (c *Client) RetryNow(ctx context.Context, req *RetryNowRequest) (*RetryNowResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.StepID, "install_workflow_steps", executeworkflowstep.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find executeworkflowstep queue signal: %w", err)
	}

	handle, err := c.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "retry-now",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				executeworkflowstep.RetryNowRequest{},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send retry-now update: %w", err)
	}

	var resp RetryNowResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("unable to get retry-now response: %w", err)
	}

	resp.WorkflowID = qs.Workflow.ID
	return &resp, nil
}
