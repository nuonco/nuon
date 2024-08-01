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
			"ID",
			"NAME",
			"PLATFORM",
			"STATUS",
			"DESCRIPTION",
		},
	}
	for _, app := range apps {
		data = append(data, []string{
			app.ID,
			app.Name,
			string(app.CloudPlatform),
			app.Status,
			app.Description,
		})
	}
	view.Render(data)
	return nil
}
