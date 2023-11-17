package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Components(ctx context.Context, installID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}
	view := ui.NewGetView()

	components, err := s.api.GetInstallComponents(ctx, installID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(components)
		return
	}

	data := [][]string{
		{
			"id",
			"name",
			"status",
			"latest deploy",
			"latest release",
		},
	}
	for _, comp := range components {
		if len(comp.InstallDeploys) != 0 {
			data = append(data, []string{
				comp.Component.ID,
				comp.Component.Name,
				comp.InstallDeploys[0].Status,
				comp.InstallDeploys[0].ID,
				comp.InstallDeploys[0].ReleaseID,
			})
		}
	}
	view.Render(data)
}
