package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// RetryGroupRequest is the input for retrying an entire workflow step group.
type RetryGroupRequest struct {
	InstallWorkflowID string
	// StepID is any step in the group to retry. The handler resolves the
	// group from this step's GroupIdx.
	StepID string
}

// RetryGroupResponse is the response from the retry-group update.
type RetryGroupResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

// RetryGroup sends a "retry-group" update to the execute-flow handler workflow.
// This retries all steps in the group that contains the given step.
func (c *Client) RetryGroup(ctx context.Context, req *RetryGroupRequest) (*RetryGroupResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	handle, err := handler.UpdateWithStart(ctx, c.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "retry-group",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			executeflow.RetryGroupRequest{
				StepID: req.StepID,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send retry-group update: %w", err)
	}

	var resp RetryGroupResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("unable to get retry-group response: %w", err)
	}

	return &resp, nil
}
