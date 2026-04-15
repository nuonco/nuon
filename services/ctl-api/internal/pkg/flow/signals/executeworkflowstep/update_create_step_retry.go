package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CreateStepRetryResponse is the response from the "create-step-retry" update handler.
type CreateStepRetryResponse struct {
	NewStepID string `json:"new_step_id"`
}

func (s *Signal) createStepRetryHandler(ctx workflow.Context) (*CreateStepRetryResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step")
	}
	if !step.Retryable {
		return nil, errors.New("step is not retryable")
	}

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get workflow")
	}

	// Mark original step as discarded
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusDiscarded,
			StatusHumanDescription: "Step was discarded and retried by the user.",
		},
	}); err != nil {
		return nil, errors.Wrap(err, "unable to mark step as discarded")
	}

	// Clone the step
	if err := s.cloneWorkflowStep(ctx, step, flw); err != nil {
		return nil, errors.Wrap(err, "unable to clone step for retry")
	}

	// Fetch updated steps to find new step ID
	steps, err := activities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, activities.GetFlowStepsRequest{
		FlowID: flw.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get flow steps")
	}

	newStep := steps[len(steps)-1]
	return &CreateStepRetryResponse{NewStepID: newStep.ID}, nil
}
