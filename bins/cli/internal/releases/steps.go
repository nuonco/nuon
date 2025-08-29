package releases

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Steps(ctx context.Context, releaseID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	steps, hasMore, err := s.listSteps(ctx, releaseID, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listSteps(ctx context.Context, releaseID string, offset, limit int) ([]*models.AppComponentReleaseStep, bool, error) {
	cmps, hasMore, err := s.api.GetReleaseSteps(ctx, releaseID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, false, err
	}
	return cmps, hasMore, nil
}
