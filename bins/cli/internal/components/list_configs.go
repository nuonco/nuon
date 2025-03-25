package components

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListConfigs(ctx context.Context, appID, compID string, asJSON bool) error {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	configs, err := s.listConfigs(ctx, compID)
	if err != nil {
		return view.Error(err)
	}

	ui.PrintJSON(configs)
	return nil
}

func (s *Service) listConfigs(ctx context.Context, compID string) ([]*models.AppComponentConfigConnection, error) {
	if !s.cfg.PaginationEnabled {
		cfgs, _, err := s.api.GetComponentConfigs(ctx, compID, &models.GetComponentConfigsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return cfgs, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppComponentConfigConnection, bool, error) {
		cmps, hasMore, err := s.api.GetComponentConfigs(ctx, compID, &models.GetComponentConfigsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return cmps, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
