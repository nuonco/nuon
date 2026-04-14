package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/executeflow"
)

// RetryStepRequest is the input for retrying a workflow step.
type RetryStepRequest struct {
	InstallWorkflowID string
	StepID            string
	Operation         string // "retry-step" or "skip-step"
}

// RetryStepResponse is the response from the retry-step update.
type RetryStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// RetryStep sends a "retry-step" update to the execute-flow handler workflow
// for the given install workflow. The handler workflow stays alive while
// retryable, so the update wakes it to retry the failed step.
func (c *Client) RetryStep(ctx context.Context, req *RetryStepRequest) (*RetryStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows")
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	handle, err := c.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "retry-step",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				executeflow.RetryStepRequest{
					StepID:    req.StepID,
					Operation: req.Operation,
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send retry-step update: %w", err)
	}

	var resp RetryStepResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("unable to get retry-step response: %w", err)
	}

	return &resp, nil
}
