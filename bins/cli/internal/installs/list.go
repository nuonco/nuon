package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, asJSON bool) {
	view := ui.NewListView()

	var (
		installs []*models.AppInstall
		err      error
	)

	if appID != "" {
		appID, err := lookup.AppID(ctx, s.api, appID)
		if err != nil {
			ui.PrintError(err)
			return
		}
		installs, err = s.api.GetAppInstalls(ctx, appID)

	} else {
		installs, err = s.api.GetAllInstalls(ctx)
	}
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(installs)
		return
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"SANDBOX",
			"RUNNER",
			"COMPONENTS",
			"CREATED AT",
		},
	}
	for _, install := range installs {
		data = append(data, []string{
			install.ID,
			install.Name,
			install.SandboxStatus,
			install.RunnerStatus,
			install.CompositeComponentStatus,
			install.CreatedAt,
		})
	}
	view.Render(data)
}
