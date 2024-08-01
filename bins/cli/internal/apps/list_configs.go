package apps

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListConfigs(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewListView()

	cfgs, err := s.api.GetAppConfigs(ctx, appID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(cfgs)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"VERSION",
			"STATUS",
			"CREATED BY",
			"CREATED AT",
		},
	}
	for _, cfg := range cfgs {
		data = append(data, []string{
			cfg.ID,
			fmt.Sprintf("%d", cfg.Version),
			string(cfg.Status),
			cfg.CreatedBy.Email,
			cfg.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}
