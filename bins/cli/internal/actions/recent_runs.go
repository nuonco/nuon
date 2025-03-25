package actions

import (
	"context"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetRecentRuns(ctx context.Context, installID, actionWorkflowID string, asJSON bool) error {
	view := ui.NewListView()

	response, err := s.getRecentRuns(ctx, installID, actionWorkflowID)

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

// GetRecentRuns fetches recent runs for an action workflow
func (s *Service) getRecentRuns(ctx context.Context, installID, actionWorkflowID string) (*models.AppInstallActionWorkflow, error) {
	if !s.cfg.PaginationEnabled {
		iaw, _, err := s.api.GetInstallActionWorkflowRecentRuns(ctx, installID, actionWorkflowID, nil)

		if err != nil {
			return nil, err
		}
		return iaw, nil
	}

	offset := 0
	pageSize := 10
	allRuns := []*models.AppInstallActionWorkflowRun{}

	for {
		iaw, hasMore, err := s.api.GetInstallActionWorkflowRecentRuns(ctx, installID, actionWorkflowID, &models.GetInstallActionWorkflowRecentRunsQuery{
			Offset:            offset,
			Limit:             pageSize,
			PaginationEnabled: true,
		})

		if err != nil {
			return nil, err
		}

		allRuns = append(allRuns, iaw.Runs...)
		iaw.Runs = allRuns

		if !hasMore {
			return iaw, nil
		}

		if len(allRuns) >= 50 {
			return iaw, nil
		}

		offset += pageSize
	}
}
