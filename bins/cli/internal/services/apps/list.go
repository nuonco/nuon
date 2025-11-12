package apps

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	apps, hasMore, err := s.listApps(ctx, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(apps)
		return nil
	}

	data := [][]string{
		{
			" NAME",
			"ID",
			"PLATFORM",
			"STATUS",
			"DESCRIPTION",
		},
	}
	curID := s.cfg.GetString("app_id")
	for _, app := range apps {
		if curID != "" {
			if app.ID == curID {
				app.Name = "*" + app.Name
			} else {
				app.Name = " " + app.Name
			}
		}
		data = append(data, []string{
			app.Name,
			app.ID,
			string(app.CloudPlatform),
			app.Status,
			app.Description,
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listApps(ctx context.Context, offset, limit int) ([]*models.AppApp, bool, error) {
	apps, hasMore, err := s.api.GetApps(ctx, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return apps, hasMore, nil
}
