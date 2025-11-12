package actions

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeleteWorkflow(ctx context.Context, action_workflow_id string) error {
	view := ui.NewDeleteView("action", action_workflow_id)
	view.Start()
	view.Update("deleting action")

	_, err := s.api.DeleteActionWorkflow(ctx, action_workflow_id)
	if err != nil {
		return view.Fail(err)
	}

	view.SuccessQueued()
	return nil
}
