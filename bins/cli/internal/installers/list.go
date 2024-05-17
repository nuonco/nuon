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
			"name",
		},
	}
	for _, installer := range installers {
		data = append(data, []string{
			installer.Metadata.Name,
		})
	}
	view.Render(data)
}
