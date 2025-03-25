package components

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appNameOrID string, asJSON bool) error {
	view := ui.NewListView()

	var (
		components []*models.AppComponent
		err        error
	)
	if appNameOrID != "" {
		appID, err := lookup.AppID(ctx, s.api, appNameOrID)
		if err != nil {
			return view.Error(err)
		}
		components, err = s.listAppComponents(ctx, appID)
	} else {
		components, err = s.listComponents(ctx)
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
	view.Render(data)
	return nil
}

func (s *Service) listComponents(ctx context.Context) ([]*models.AppComponent, error) {
	if !s.cfg.PaginationEnabled {
		cmps, _, err := s.api.GetAllComponents(ctx, &models.GetAllComponentsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return cmps, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppComponent, bool, error) {
		cmps, hasMore, err := s.api.GetAllComponents(ctx, &models.GetAllComponentsQuery{
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

func (s *Service) listAppComponents(ctx context.Context, appID string) ([]*models.AppComponent, error) {
	if !s.cfg.PaginationEnabled {
		cmps, _, err := s.api.GetAppComponents(ctx, appID, &models.GetAppComponentsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return cmps, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppComponent, bool, error) {
		cmps, hasMore, err := s.api.GetAppComponents(ctx, appID, &models.GetAppComponentsQuery{
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
