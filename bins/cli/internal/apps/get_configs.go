package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetConfigs(ctx context.Context, appID string, asJSON bool) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	cfgs, err := s.api.GetAppConfigs(ctx, appID)
	if err != nil {
		view.Error(err)
		return
	}

	data := [][]string{
		{
			"id",
			"name",
			"status",
		},
	}
	for _, cfg := range cfgs {
		data = append(data, []string{
			cfg.ID,
			string(cfg.Status),
			cfg.CreatedAt,
		})
	}
	view.Render(data)

}
