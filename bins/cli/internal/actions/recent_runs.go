package actions

import (
	"context"
	"time"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetRecentRuns(ctx context.Context, installID, actionWorkflowID string, asJSON bool) error {
	view := ui.NewListView()

	response, err := s.api.GetInstallActionWorkflowRecentRuns(ctx, installID, actionWorkflowID)

	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(response.Runs)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"Trigger Type",
			"Status",
			"Status Description",
			"Execution Time",
		},
	}

	for _, run := range response.Runs {
		data = append(data, []string{
			run.ID,
			string(run.TriggerType),
			run.Status,
			run.StatusDescription,
			time.Duration(run.ExecutionTime).String(),
		})
	}
	view.Render(data)
	return nil
}
