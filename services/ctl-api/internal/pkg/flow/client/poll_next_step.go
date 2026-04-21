package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// PollNextStepRequest identifies the workflow to poll.
type PollNextStepRequest struct {
	InstallWorkflowID string
}

// PollNextStepResponse contains the current in-flight step, or empty fields
// when all steps are terminal (workflow complete).
type PollNextStepResponse struct {
	StepID  string `json:"step_id"`
	StepIdx int    `json:"step_idx"`
	Status  string `json:"status"`
}

// PollNextStep sends a "poll-next-step" update to the execute-flow handler
// workflow. It returns the first non-terminal step, or an empty response when
// the workflow is complete.
func (c *Client) PollNextStep(ctx context.Context, req *PollNextStepRequest) (*PollNextStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	handle, err := c.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "poll-next-step",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send poll-next-step update: %w", err)
	}

	var flowResp executeflow.PollNextStepResponse
	if err := handle.Get(ctx, &flowResp); err != nil {
		return nil, fmt.Errorf("unable to get poll-next-step response: %w", err)
	}

	return &PollNextStepResponse{
		StepID:  flowResp.StepID,
		StepIdx: flowResp.StepIdx,
		Status:  string(flowResp.Status),
	}, nil
}
