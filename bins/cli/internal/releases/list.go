package releases

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID, compID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	var (
		hasMore  bool
		releases []*models.AppComponentRelease
		err      error
	)

	if compID != "" {
		compID, err = lookup.ComponentID(ctx, s.api, appID, compID)
		if err != nil {
			return ui.PrintError(err)
		}

		releases, hasMore, err = s.listComponentReleases(ctx, compID, offset, limit)
	} else if appID != "" {
		appID, err = lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}

		releases, hasMore, err = s.listAppComponentReleases(ctx, appID, offset, limit)
	}
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(releases)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"STATUS",
			"BUILD ID",
			"CREATED AT",
		},
	}
	for _, release := range releases {
		data = append(data, []string{
			release.ID,
			release.Status,
			release.BuildID,
			release.CreatedAt,
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listComponentReleases(ctx context.Context, compID string, offset, limit int) ([]*models.AppComponentRelease, bool, error) {
	cmps, hasMore, err := s.api.GetComponentReleases(ctx, compID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, false, err
	}
	return cmps, hasMore, nil
}

func (s *Service) listAppComponentReleases(ctx context.Context, appID string, offset, limit int) ([]*models.AppComponentRelease, bool, error) {
	cmps, hasMore, err := s.api.GetAppReleases(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return cmps, hasMore, nil
}
