package installers

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) {
	view := ui.NewListView()

	installers, err := s.api.GetInstallers(ctx)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(installers)
		return
	}

	data := [][]string{
		{
			"id",
			"name",
			"created at",
		},
	}
	for _, installer := range installers {
		data = append(data, []string{
			installer.ID,
			installer.Metadata.Name,
			installer.CreatedAt,
		})
	}
	view.Render(data)
}
