package apps

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) error {
	view := ui.NewListView()

	apps, err := s.listApps(ctx)
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
	view.Render(data)
	return nil
}

func (s *Service) listApps(ctx context.Context) ([]*models.AppApp, error) {
	if !s.cfg.PaginationEnabled {
		apps, _, err := s.api.GetApps(ctx, nil)
		if err != nil {
			return nil, err
		}
		return apps, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppApp, bool, error) {
		apps, hasMore, err := s.api.GetApps(ctx, &models.GetAppsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: true,
		})
		if err != nil {
			return nil, false, err
		}
		return apps, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
