package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// groupFinishedHandler blocks until Execute() completes, then reads the
// group's ResultDirective and returns it. When the group has a StepGroupID,
// reads from the step group record directly. Falls back to reading the
// workflow's ResultDirective for backward compatibility with synthetic groups.
func (s *Signal) groupFinishedHandler(ctx workflow.Context) (*activities.GroupFinishedResponse, error) {
	group, err := activities.AwaitPkgWorkflowsFlowGetFlowStepGroupByID(ctx, s.StepGroupID)
	if err != nil {
		return nil, err
	}

	if generics.SliceContains(group.Status.Status, []app.Status{
		app.StatusError,
		app.StatusCancelled,
		app.StatusSuccess,
	}) {
		return &activities.GroupFinishedResponse{Directive: group.ResultDirective}, nil
	}

	// NOTE(jm): if a workflow signal failed / panicked and was in flight, this will wait forever and block. We are
	// fast following with some improvements to this.
	if err := workflow.Await(ctx, func() bool { return s.finished }); err != nil {
		return nil, err
	}

	group, err = activities.AwaitPkgWorkflowsFlowGetFlowStepGroupByID(ctx, s.StepGroupID)
	if err != nil {
		return nil, err
	}

	return &activities.GroupFinishedResponse{Directive: group.ResultDirective}, nil
}
