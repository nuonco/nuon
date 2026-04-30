package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// stepFinishedHandler blocks until Execute() completes, then fetches the step
// from the database and returns its final status and directive. This provides
// callers (the group signal) with a resilient way to get the step result even
// after handler termination and restart.
func (s *Signal) stepFinishedHandler(ctx workflow.Context) (*activities.StepFinishedResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return nil, err
	}
	if generics.SliceContains(step.Status.Status, []app.Status{
		app.StatusError,
		app.StatusCancelled,
		app.StatusSuccess,
	}) {
		return &activities.StepFinishedResponse{
			StepID:    step.ID,
			Status:    step.Status.Status,
			Directive: step.ResultDirective,
		}, nil
	}

	// (jm): if a workflow signal failed / panicked and was in flight, this will wait forever and block. We are
	// fast following with some improvements to this.
	if err := workflow.Await(ctx, func() bool { return s.finished }); err != nil {
		return nil, err
	}

	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return nil, err
	}

	return &activities.StepFinishedResponse{
		StepID:    step.ID,
		Status:    step.Status.Status,
		Directive: step.ResultDirective,
	}, nil
}
