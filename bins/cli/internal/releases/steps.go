package releases

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Steps(ctx context.Context, releaseID string) {
	view := ui.NewListView()

	steps, err := s.api.GetReleaseSteps(ctx, releaseID)
	if err != nil {
		view.Error(err)
		return
	}
	data := [][]string{
		{
			"id",
			"status",
			"created at",
			"delay",
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
}
