package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// ApprovePlanRequest is the input for approving a plan on a workflow step.
type ApprovePlanRequest struct {
	InstallWorkflowID  string
	StepID             string
	ApprovalResponseID string
	ResponseType       app.WorkflowStepResponseType
}

// ApprovePlan sends an "approve-step" update to the execute-flow handler workflow
// for the given install workflow. The execute-flow handler forwards the approval to
// the step's handler workflow.
func (c *Client) ApprovePlan(ctx context.Context, req *ApprovePlanRequest) error {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "", executeflow.SignalType)
	if err != nil {
		return fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	handle, err := handler.UpdateWithStart(ctx, c.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "approve-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			executeflow.ApproveStepRequest{
				StepID:             req.StepID,
				ApprovalResponseID: req.ApprovalResponseID,
				ResponseType:       string(req.ResponseType),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("unable to send approve-step update: %w", err)
	}

	var resp executeflow.ApproveStepResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return fmt.Errorf("approve-step update failed: %w", err)
	}

	return nil
}
