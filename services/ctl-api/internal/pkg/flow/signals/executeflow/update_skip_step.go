package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// SkipStepRequest is the input for the "skip-step" update handler.
type SkipStepRequest struct {
	StepID string `json:"step_id"`
}

// SkipStepResponse is the response from the "skip-step" update handler.
type SkipStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Skippable  bool   `json:"skippable"`
}

func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	// Steps awaiting approval can be skipped via the approval mechanism
	if step.Status.Status == app.AwaitingApproval || step.Status.Status == app.Status("awaiting-approval") {
		if _, err := workflowactivities.AwaitForwardApprovePlan(ctx, workflowactivities.ForwardApprovePlanRequest{
			StepID:       req.StepID,
			ResponseType: string(app.WorkflowStepApprovalResponseTypeSkipCurrent),
		}); err != nil {
			return nil, fmt.Errorf("unable to forward skip approval for step %s: %w", req.StepID, err)
		}

		s.resumeRequested = true
		s.resumeRunType = app.WorkflowRunTypeSkip
		s.resumeStepID = req.StepID
		return &SkipStepResponse{
			WorkflowID: s.WorkflowID,
			Skippable:  true,
		}, nil
	}

	// Only error state steps can be skipped via the direct path
	if step.Status.Status != app.StatusError {
		return &SkipStepResponse{
			WorkflowID: s.WorkflowID,
			Skippable:  false,
		}, nil
	}

	if !step.Skippable {
		return &SkipStepResponse{
			WorkflowID: s.WorkflowID,
			Skippable:  false,
		}, nil
	}

	// Mark the errored step as skipped so the conductor advances past it.
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: req.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusUserSkipped,
			StatusHumanDescription: "Step was skipped by the user.",
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to mark step %s as skipped: %w", req.StepID, err)
	}

	s.resumeRequested = true
	s.resumeRunType = app.WorkflowRunTypeSkip
	s.resumeStepID = req.StepID
	return &SkipStepResponse{
		WorkflowID: s.WorkflowID,
		Skippable:  true,
	}, nil
}
