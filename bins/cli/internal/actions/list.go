package actions

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	if appID == "" {
		s.printAppNotSetMsg()
		return nil
	}

	wfs, hasMore, err := s.getActionWorkflows(ctx, appID, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) getActionWorkflows(ctx context.Context, appID string, offset, limit int) ([]*models.AppActionWorkflow, bool, error) {
	wfs, hasMore, err := s.api.GetActionWorkflows(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return wfs, hasMore, nil
}
