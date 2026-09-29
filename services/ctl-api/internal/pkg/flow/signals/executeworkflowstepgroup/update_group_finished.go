package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

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

	if err := workflow.Await(ctx, func() bool { return s.finished }); err != nil {
		return nil, err
	}

	group, err = activities.AwaitPkgWorkflowsFlowGetFlowStepGroupByID(ctx, s.StepGroupID)
	if err != nil {
		return nil, err
	}

	return &activities.GroupFinishedResponse{Directive: group.ResultDirective}, nil
}
