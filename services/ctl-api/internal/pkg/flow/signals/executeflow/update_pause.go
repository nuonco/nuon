package executeflow

import "go.temporal.io/sdk/workflow"

// pauseWorkflowHandler sets the pause flag so the flow pauses after
// the current group completes.
func (s *Signal) pauseWorkflowHandler(ctx workflow.Context) error {
	s.pauseRequested = true
	return nil
}
