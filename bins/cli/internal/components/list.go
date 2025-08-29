package components

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appNameOrID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	var (
		components []*models.AppComponent
		err        error
		hasMore    bool
	)
	if appNameOrID != "" {
		appID, err := lookup.AppID(ctx, s.api, appNameOrID)
		if err != nil {
			return view.Error(err)
		}
		components, hasMore, err = s.listAppComponents(ctx, appID, offset, limit)
	} else {
		components, hasMore, err = s.listComponents(ctx, offset, limit)
	}
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(components)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"CREATED AT",
			"UPDATED AT",
			"CREATED BY",
			"CONFIG VERSIONS",
		},
	}
	for _, component := range components {
		data = append(data, []string{
			component.ID,
			component.Name,
			component.CreatedAt,
			component.UpdatedAt,
			component.CreatedByID,
			strconv.Itoa(int(component.ConfigVersions)),
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listComponents(ctx context.Context, offset, limit int) ([]*models.AppComponent, bool, error) {
	cmps, hasMore, err := s.api.GetAllComponents(ctx, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return cmps, hasMore, nil
}

func (s *Service) listAppComponents(ctx context.Context, appID string, offset, limit int) ([]*models.AppComponent, bool, error) {
	cmps, hasMore, err := s.api.GetAppComponents(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return cmps, hasMore, nil
}
