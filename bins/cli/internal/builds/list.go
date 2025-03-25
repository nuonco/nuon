package builds

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, compID, appID string, limit *int, asJSON bool) error {
	var err error
	if compID != "" {
		compID, err = lookup.ComponentID(ctx, s.api, appID, compID)
		if err != nil {
			return ui.PrintError(err)
		}
	}
	if appID != "" {
		appID, err = lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}
	}

	view := ui.NewListView()

	builds, err := s.listComponentBuilds(ctx, compID, appID, limit)
	if err != nil {
		return view.Error(errors.Wrap(err, "failed to fetch component builds"))
	}

	if asJSON {
		ui.PrintJSON(builds)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"STATUS",
			"COMPONENT NAME",
			"CONFIG VERSION",
			"GIT REF / BRANCH",
			"CREATED AT",
		},
	}
	for _, build := range builds {
		data = append(data, []string{
			build.ID,
			build.Status,
			build.ComponentName,
			fmt.Sprintf("%d", build.ComponentConfigVersion),
			build.GitRef,
			build.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) listComponentBuilds(ctx context.Context, compID, appID string, limit *int) ([]*models.AppComponentBuild, error) {
	limitInt := 10
	if limit != nil {
		limitInt = *limit
	}

	if !s.cfg.PaginationEnabled {
		builds, _, err := s.api.GetComponentBuilds(ctx, compID, appID, &models.GetComponentBuildsQuery{
			Offset:            0,
			Limit:             limitInt,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return builds, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppComponentBuild, bool, error) {
		builds, hasMore, err := s.api.GetComponentBuilds(ctx, compID, appID, &models.GetComponentBuildsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return builds, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, limitInt, fetchFn)
}
