package installs

import (
	"context"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Workflows(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	workflows, hasMore, err := s.listWorkflows(ctx, installID, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(workflows)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"TYPE",
			"STATUS",
			"STARTED AT",
			"FINISHED AT",
			"UPDATED AT",
		},
	}
	for _, workflow := range workflows {
		startedAt, _ := time.Parse(time.RFC3339Nano, workflow.StartedAt)
		finishedAt, _ := time.Parse(time.RFC3339Nano, workflow.FinishedAt)
		updatedAt, _ := time.Parse(time.RFC3339Nano, workflow.UpdatedAt)
		status := ""
		if workflow.Status != nil {
			status = string(workflow.Status.Status)
		}

		data = append(data, []string{
			workflow.ID,
			workflow.Name,
			string(workflow.Type),
			status,
			startedAt.Format(time.Stamp),
			finishedAt.Format(time.Stamp),
			updatedAt.Format(time.Stamp),
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listWorkflows(ctx context.Context, appID string, offset, limit int) ([]*models.AppWorkflow, bool, error) {
	workflows, hasMore, err := s.api.GetWorkflows(ctx, appID, &models.GetPaginatedQuery{
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return workflows, hasMore, nil
}
