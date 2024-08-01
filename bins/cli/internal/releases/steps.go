package releases

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Steps(ctx context.Context, releaseID string, asJSON bool) error {
	view := ui.NewListView()

	steps, err := s.api.GetReleaseSteps(ctx, releaseID)
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
