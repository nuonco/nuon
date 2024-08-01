package installers

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) error {
	view := ui.NewListView()

	installers, err := s.api.GetInstallers(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(installers)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"CREATED AT",
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
	return nil
}
