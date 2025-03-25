package releases

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Steps(ctx context.Context, releaseID string, asJSON bool) error {
	view := ui.NewListView()

	steps, err := s.listSteps(ctx, releaseID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(steps)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"STATUS",
			"CREATED AT",
			"DELAY",
		},
	}
	for _, step := range steps {
		data = append(data, []string{
			step.ID,
			step.Status,
			step.CreatedAt,
			step.Delay,
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) listSteps(ctx context.Context, releaseID string) ([]*models.AppComponentReleaseStep, error) {
	if !s.cfg.PaginationEnabled {
		releases, _, err := s.api.GetReleaseSteps(ctx, releaseID, &models.GetReleaseStepsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return releases, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppComponentReleaseStep, bool, error) {
		cmps, hasMore, err := s.api.GetReleaseSteps(ctx, releaseID, &models.GetReleaseStepsQuery{
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
