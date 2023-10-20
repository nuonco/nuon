package installs

import (
	"context"
	"strconv"

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
			"created at",
			"updated at",
			"created by",
			"config versions",
		},
	}
	for _, comp := range components {
		data = append(data, []string{
			comp.Component.ID,
			comp.Component.Name,
			comp.Component.CreatedAt,
			comp.Component.UpdatedAt,
			comp.Component.CreatedByID,
			strconv.Itoa(int(comp.Component.ConfigVersions)),
		})
	}
	view.Render(data)
}
