package actions

import (
	"context"
	"time"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetRun(ctx context.Context, installID, runID string, asJSON bool) error {
	if runID == "" {
		return nil
	}

	view := ui.NewGetView()

	run, err := s.api.GetInstallActionWorkflowRun(ctx, installID, runID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(run)
		return nil
	}

	view.Render([][]string{
		{"id", run.ID},
		{"trigger", string(run.TriggerType)},
		{"status", run.Status},
		{"status description", run.StatusDescription},
		{"execution time", time.Duration(run.ExecutionTime).String()},
		{"created at", run.CreatedAt},
		{"updated at", run.UpdatedAt},
		{"created by", run.CreatedByID},
	})

	return nil
}
