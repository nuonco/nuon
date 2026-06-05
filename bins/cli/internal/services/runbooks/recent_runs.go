package runbooks

import (
	"context"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) GetRecentRuns(ctx context.Context, installID, runbookID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	runs, hasMore, err := s.api.GetInstallRunbookRuns(ctx, installID, runbookID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(runs)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"RUNBOOK",
			"STATUS",
			"STATUS DESCRIPTION",
			"EXECUTION TIME",
			"CREATED AT",
		},
	}

	for _, run := range runs {
		runbookName := ""
		if run.InstallRunbook != nil && run.InstallRunbook.Runbook != nil {
			runbookName = run.InstallRunbook.Runbook.Name
		}
		data = append(data, []string{
			run.ID,
			runbookName,
			run.Status,
			run.StatusDescription,
			time.Duration(run.ExecutionTime).String(),
			run.CreatedAt,
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}
