package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewListView()

	var (
		installs []*models.AppInstall
		err      error
	)

	if appID != "" {
		appID, err := lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}
		installs, err = s.api.GetAppInstalls(ctx, appID)

	} else {
		installs, err = s.api.GetAllInstalls(ctx)
	}
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(installs)
		return nil
	}

	data := [][]string{
		{
			"NAME",
			"ID",
			"SANDBOX",
			"RUNNER",
			"COMPONENTS",
			"CREATED AT",
		},
	}
	curID := s.cfg.GetString("org_id")
	for _, install := range installs {
		if curID != "" {
			if install.ID == curID {
				install.Name = "*" + install.Name
			} else {
				install.Name = " " + install.Name
			}
		}
		data = append(data, []string{
			install.Name,
			install.ID,
			install.SandboxStatus,
			install.RunnerStatus,
			install.CompositeComponentStatus,
			install.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}
