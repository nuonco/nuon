package actions

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewListView()

	if appID == "" {
		s.printAppNotSetMsg()
		return nil
	}

	wfs, err := s.getActionWorkflows(ctx, appID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(wfs)
		return nil
	}

	data := [][]string{
		{
			"NAME",
			"ID",
		},
	}

	for _, wf := range wfs {
		data = append(data, []string{
			wf.Name,
			wf.ID,
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) getActionWorkflows(ctx context.Context, appID string) ([]*models.AppActionWorkflow, error) {
	if !s.cfg.PaginationEnabled {
		wfs, _, err := s.api.GetActionWorkflows(ctx, appID, nil)
		if err != nil {
			return nil, err
		}
		return wfs, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppActionWorkflow, bool, error) {
		wfs, hasMore, err := s.api.GetActionWorkflows(ctx, appID, &models.GetActionWorkflowsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: true,
		})
		if err != nil {
			return nil, false, err
		}

		return wfs, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
