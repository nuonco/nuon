package runbooks

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) CreateRun(ctx context.Context, installID, runbookID string, asJSON bool) error {
	view := ui.NewGetView()

	run, err := s.api.CreateInstallRunbookRun(ctx, installID, runbookID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(run)
		return nil
	}

	view.Render([][]string{
		{"id", run.ID},
		{"status", run.Status},
		{"status description", run.StatusDescription},
		{"install workflow id", run.InstallWorkflowID},
		{"created at", run.CreatedAt},
	})

	return nil
}
