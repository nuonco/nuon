package executeflow

import "go.temporal.io/sdk/workflow"

func (s *Signal) pauseWorkflowHandler(ctx workflow.Context) error {
	defer s.beginUpdate()()

	s.pauseRequested = true
	return nil
}
