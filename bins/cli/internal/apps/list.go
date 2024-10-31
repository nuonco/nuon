package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) error {
	view := ui.NewListView()

	apps, err := s.api.GetApps(ctx)
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
