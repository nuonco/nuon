package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	app, err := s.api.GetApp(ctx, appID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(app)
		return nil
	}

	view.Render([][]string{
		{"id", app.ID},
		{"name", app.Name},
		{"platform", string(app.CloudPlatform)},
		{"status", app.Status},
		{"created at", app.CreatedAt},
		{"updated at", app.UpdatedAt},
		{"created by", app.CreatedByID},
	})

	return nil
}
