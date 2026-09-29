package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type PollNextStepRequest struct {
	InstallWorkflowID string
}

type PollNextStepResponse struct {
	StepID  string `json:"step_id"`
	StepIdx int    `json:"step_idx"`
	Status  string `json:"status"`
}

func (c *Client) PollNextStep(ctx context.Context, req *PollNextStepRequest) (*PollNextStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	var flowResp executeflow.PollNextStepResponse
	if err := c.updateWithStartUntilCompleted(ctx, qs, "poll-next-step", &flowResp); err != nil {
		return nil, fmt.Errorf("unable to get poll-next-step response: %w", err)
	}

	return &PollNextStepResponse{
		StepID:  flowResp.StepID,
		StepIdx: flowResp.StepIdx,
		Status:  string(flowResp.Status),
	}, nil
}
