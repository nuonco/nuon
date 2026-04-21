package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"

	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// groupFinishedHandler blocks until Execute() completes, then reads the
// workflow's ResultDirective and returns it. This provides callers (the flow
// signal) with a resilient way to get the group result even after handler
// termination and restart.
func (s *Signal) groupFinishedHandler(ctx workflow.Context) (*activities.GroupFinishedResponse, error) {
	if err := workflow.Await(ctx, func() bool { return s.finished }); err != nil {
		return nil, err
	}

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return nil, err
	}

	return &activities.GroupFinishedResponse{Directive: flw.ResultDirective}, nil
}
