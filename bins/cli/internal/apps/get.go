package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, appID string, asJSON bool) {
	view := ui.NewGetView()

	app, err := s.api.GetApp(ctx, appID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(app)
		return
	}

	view.Render([][]string{
		{"id", app.ID},
		{"name", app.Name},
		{"status", app.Status},
		{"created at", app.CreatedAt},
		{"updated at", app.UpdatedAt},
		{"created by", app.CreatedByID},
	})
}
