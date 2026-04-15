package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// SkipStepRequest is the input for skipping a workflow step.
type SkipStepRequest struct {
	InstallWorkflowID string
	StepID            string
}

// SkipStepResponse is the response from the skip-step update.
type SkipStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Skippable  bool   `json:"skippable"`
}

// SkipStep sends a "skip-step" update to the execute-flow handler workflow
// for the given install workflow.
func (c *Client) SkipStep(ctx context.Context, req *SkipStepRequest) (*SkipStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	handle, err := c.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "skip-step",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				executeflow.SkipStepRequest{
					StepID: req.StepID,
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send skip-step update: %w", err)
	}

	var resp SkipStepResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("unable to get skip-step response: %w", err)
	}

	return &resp, nil
}
