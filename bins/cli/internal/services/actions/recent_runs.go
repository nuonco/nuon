package actions

import (
	"context"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetRecentRuns(ctx context.Context, installID, actionWorkflowID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	response, hasMore, err := s.getRecentRuns(ctx, installID, actionWorkflowID, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

// GetRecentRuns fetches recent runs for an action workflow
func (s *Service) getRecentRuns(ctx context.Context, installID, actionWorkflowID string, offset, limit int) (*models.AppInstallActionWorkflow, bool, error) {
	iaw, hasMore, err := s.api.GetInstallActionWorkflowRecentRuns(ctx, installID, actionWorkflowID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return iaw, hasMore, nil
}
