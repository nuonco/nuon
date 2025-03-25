package releases

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID, compID string, asJSON bool) error {
	view := ui.NewListView()

	var (
		releases []*models.AppComponentRelease
		err      error
	)

	if compID != "" {
		compID, err = lookup.ComponentID(ctx, s.api, appID, compID)
		if err != nil {
			return ui.PrintError(err)
		}

		releases, err = s.listComponentReleases(ctx, compID)
	} else if appID != "" {
		appID, err = lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}

		releases, err = s.listAppComponentReleases(ctx, appID)
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
	view.Render(data)
	return nil
}

func (s *Service) listComponentReleases(ctx context.Context, compID string) ([]*models.AppComponentRelease, error) {
	if !s.cfg.PaginationEnabled {
		releases, _, err := s.api.GetComponentReleases(ctx, compID, &models.GetComponentReleasesQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return releases, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppComponentRelease, bool, error) {
		cmps, hasMore, err := s.api.GetComponentReleases(ctx, compID, &models.GetComponentReleasesQuery{
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

func (s *Service) listAppComponentReleases(ctx context.Context, appID string) ([]*models.AppComponentRelease, error) {
	if !s.cfg.PaginationEnabled {
		cmps, _, err := s.api.GetAppReleases(ctx, appID, &models.GetAppReleasesQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return cmps, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppComponentRelease, bool, error) {
		cmps, hasMore, err := s.api.GetAppReleases(ctx, appID, &models.GetAppReleasesQuery{
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
