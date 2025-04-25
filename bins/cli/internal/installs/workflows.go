package installs

import (
	"context"
	"time"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Workflows(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	workflows, err := s.listWorkflows(ctx, installID)
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
	view.Render(data)
	return nil
}

func (s *Service) listWorkflows(ctx context.Context, appID string) ([]*models.AppInstallWorkflow, error) {
	if !s.cfg.PaginationEnabled {
		workflows, _, err := s.api.GetInstallWorkflows(ctx, appID, &models.GetInstallWorkflowsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return workflows, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppInstallWorkflow, bool, error) {
		workflows, hasMore, err := s.api.GetInstallWorkflows(ctx, appID, &models.GetInstallWorkflowsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return workflows, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
