package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context) {
	view := ui.NewListView()

	apps, err := s.api.GetApps(ctx)
	if err != nil {
		view.Error(err)
	}

	data := [][]string{
		[]string{
			"id",
			"name",
			"status",
		},
	}
	for _, app := range apps {
		data = append(data, []string{
			app.ID,
			app.Name,
			app.Status,
		})
	}
	view.Render(data)
}
