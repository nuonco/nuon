package executeflow

import "go.temporal.io/sdk/workflow"

func (s *Signal) unpauseWorkflowHandler(ctx workflow.Context) error {
	defer s.beginUpdate()()

	s.pauseRequested = false
	s.resumeRequested = true
	return nil
}
