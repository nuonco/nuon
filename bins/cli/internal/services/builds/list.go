package builds

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, compID, appID string, offset, limit int, asJSON bool) error {
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

	builds, hasMore, err := s.listComponentBuilds(ctx, compID, appID, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listComponentBuilds(ctx context.Context, compID, appID string, offset, limit int) ([]*models.AppComponentBuild, bool, error) {
	builds, hasMore, err := s.api.GetComponentBuilds(ctx, compID, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return builds, hasMore, nil
}
