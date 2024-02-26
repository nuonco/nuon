package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListConfigs(ctx context.Context, appID string, asJSON bool) {
	view := ui.NewListView()

	cfgs, err := s.api.GetAppConfigs(ctx, appID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(cfgs)
		return
	}

	data := [][]string{
		{
			"id",
			"status",
			"description",
			"created at",
		},
	}
	for _, app := range cfgs {
		data = append(data, []string{
			app.ID,
			string(app.Status),
			app.StatusDescription,
			app.CreatedAt,
		})
	}
	view.Render(data)
}
