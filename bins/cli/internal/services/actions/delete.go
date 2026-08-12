package actions

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) DeleteWorkflow(ctx context.Context, actionWorkflowID string, asJSON bool) error {
	if asJSON {
		if _, err := s.api.DeleteActionWorkflow(ctx, actionWorkflowID); err != nil {
			return err
		}
		ui.PrintJSON(map[string]string{
			"id":      actionWorkflowID,
			"status":  "queued_for_deletion",
			"message": "action workflow queued for deletion",
		})
		return nil
	}

	view := ui.NewDeleteView("action", actionWorkflowID, s.cfg.Interactive)
	view.Start()

	if _, err := s.api.DeleteActionWorkflow(ctx, actionWorkflowID); err != nil {
		return view.Fail(err)
	}

	view.SuccessQueued()
	return nil
}
