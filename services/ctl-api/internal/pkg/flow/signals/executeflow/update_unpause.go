package executeflow

import "go.temporal.io/sdk/workflow"

// unpauseWorkflowHandler clears the pause flag and triggers a resume so the
// flow continues from the next group.
func (s *Signal) unpauseWorkflowHandler(ctx workflow.Context) error {
	s.pauseRequested = false
	s.resumeRequested = true
	return nil
}
