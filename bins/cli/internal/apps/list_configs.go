package apps

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListConfigs(ctx context.Context, appID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	if appID == "" {
		s.printAppNotSetMsg()
		return nil
	}

	cfgs, hasMore, err := s.listConfigs(ctx, appID, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listConfigs(ctx context.Context, appID string, offset, limit int) ([]*models.AppAppConfig, bool, error) {
	cfgs, hasMore, err := s.api.GetAppConfigs(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return cfgs, hasMore, nil
}
