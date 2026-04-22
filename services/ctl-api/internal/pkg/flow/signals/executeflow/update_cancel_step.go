package executeflow

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CancelStepRequest is the input for the "cancel-step" update handler.
type CancelStepRequest struct {
	StepID string `json:"step_id"`
}

// CancelStepResponse is the response from the "cancel-step" update handler.
type CancelStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// cancelStepHandler sets cancelRequested immediately and propagates the
// cancellation to the group asynchronously. Unlike retry, cancellation does
// not need to wait for a response — it is fire-and-forget.
func (s *Signal) cancelStepHandler(ctx workflow.Context, req CancelStepRequest) (*CancelStepResponse, error) {
	s.cancelRequested = true

	l, _ := log.WorkflowLogger(ctx)

	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		if l != nil {
			l.Warn("cancel-step: unable to get step",
				zap.String("step_id", req.StepID),
				zap.Error(err))
		}
		return &CancelStepResponse{WorkflowID: s.WorkflowID}, nil
	}

	if _, err := workflowactivities.AwaitForwardCancelStepToGroup(ctx, workflowactivities.ForwardCancelStepToGroupRequest{
		StepID:      req.StepID,
		StepGroupID: step.WorkflowStepGroupID,
	}); err != nil {
		if l != nil {
			l.Warn("cancel-step: unable to forward cancel to group",
				zap.String("step_id", req.StepID),
				zap.Error(err))
		}
	}

	return &CancelStepResponse{WorkflowID: s.WorkflowID}, nil
}
