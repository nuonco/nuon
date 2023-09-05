package apps

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) List(ctx context.Context) error {
	apps, err := s.api.GetApps(ctx)
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		ui.Line(ctx, "No apps found")
	} else {
		for _, app := range apps {
			ui.Line(ctx, "%s - %s", app.ID, app.Name)
		}
	}

	return nil
}
