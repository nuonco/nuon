package apps

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Get(ctx context.Context, appID string) error {
	app, err := s.api.GetApp(ctx, appID)
	if err != nil {
		return err
	}

	ui.Line(ctx, "%s - %s", app.ID, app.Name)
	return nil
}
