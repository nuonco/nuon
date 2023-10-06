package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, asJSON bool) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewListView()

	var (
		installs []*models.AppInstall
	)

	if appID == "" {
		installs, err = s.api.GetAllInstalls(ctx)
	} else {
		installs, err = s.api.GetAppInstalls(ctx, appID)
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
			"id",
			"name",
			"status",
			"created at",
		},
	}
	for _, install := range installs {
		data = append(data, []string{
			install.ID,
			install.Name,
			install.Status,
			install.CreatedAt,
		})
	}
	view.Render(data)
}
