package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"

	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// groupFinishedHandler blocks until Execute() completes, then reads the
// group's ResultDirective and returns it. When the group has a StepGroupID,
// reads from the step group record directly. Falls back to reading the
// workflow's ResultDirective for backward compatibility with synthetic groups.
func (s *Signal) groupFinishedHandler(ctx workflow.Context) (*activities.GroupFinishedResponse, error) {
	if err := workflow.Await(ctx, func() bool { return s.finished }); err != nil {
		return nil, err
	}

	if s.StepGroupID != "" {
		group, err := activities.AwaitPkgWorkflowsFlowGetFlowStepGroupByID(ctx, s.StepGroupID)
		if err != nil {
			return nil, err
		}
		return &activities.GroupFinishedResponse{Directive: group.ResultDirective}, nil
	}

	// Fallback: read from workflow for backward compat with synthetic groups.
	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return nil, err
	}
	return &activities.GroupFinishedResponse{Directive: flw.ResultDirective}, nil
}
