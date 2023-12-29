package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListDeploys(ctx context.Context, installID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	deploys, err := s.api.GetInstallDeploys(ctx, installID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(deploys)
		return
	}

	data := [][]string{
		{
			"id",
			"status",
			"build id",
			"created at",
			"component id",
			"component name",
		},
	}
	for _, deploy := range deploys {
		data = append(data, []string{
			deploy.ID,
			deploy.Status,
			deploy.BuildID,
			deploy.CreatedAt,
			deploy.ComponentID,
			deploy.ComponentName,
		})
	}
	view.Render(data)
}
