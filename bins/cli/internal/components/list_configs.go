package components

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListConfigs(ctx context.Context, appID, compID string, offset, limit int, asJSON bool) error {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	configs, _, err := s.listConfigs(ctx, compID, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	ui.PrintJSON(configs)
	return nil
}

func (s *Service) listConfigs(ctx context.Context, compID string, offset, limit int) ([]*models.AppComponentConfigConnection, bool, error) {
	cmps, hasMore, err := s.api.GetComponentConfigs(ctx, compID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return cmps, hasMore, nil
}
