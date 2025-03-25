package apps

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListConfigs(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewListView()

	if appID == "" {
		s.printAppNotSetMsg()
		return nil
	}

	cfgs, err := s.listConfigs(ctx, appID)
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
			cfg.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) listConfigs(ctx context.Context, appID string) ([]*models.AppAppConfig, error) {
	if !s.cfg.PaginationEnabled {
		cfgs, _, err := s.api.GetAppConfigs(ctx, appID, nil)
		if err != nil {
			return nil, err
		}
		return cfgs, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppAppConfig, bool, error) {
		cfgs, hasMore, err := s.api.GetAppConfigs(ctx, appID, &models.GetAppConfigsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return cfgs, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
