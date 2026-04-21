package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// CancelGroupRequest is the input for cancelling all steps in a group.
type CancelGroupRequest struct {
	InstallWorkflowID string
	// StepID is any step in the group to cancel.
	StepID string
}

// CancelGroupResponse is the response from the cancel-group update.
type CancelGroupResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// CancelGroup sends a "cancel-group" update to the execute-flow handler workflow.
// This cancels all in-flight steps in the group containing the given step.
func (c *Client) CancelGroup(ctx context.Context, req *CancelGroupRequest) (*CancelGroupResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	handle, err := c.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "cancel-group",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				executeflow.CancelGroupRequest{
					StepID: req.StepID,
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send cancel-group update: %w", err)
	}

	var resp CancelGroupResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("unable to get cancel-group response: %w", err)
	}

	return &resp, nil
}
