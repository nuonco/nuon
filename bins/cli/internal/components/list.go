package components

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) List(ctx context.Context, appID string) error {
	components := []*models.AppComponent{}
	err := error(nil)
	if appID != "" {
		components, err = s.api.GetAppComponents(ctx, appID)
	} else {
		components, err = s.api.GetAllComponents(ctx)
	}
	if err != nil {
		return err
	}

	if len(components) == 0 {
		ui.Line(ctx, "No components found")
	} else {
		for _, component := range components {
			ui.Line(ctx, "%s - %s", component.ID, component.Name)
		}
	}

	return nil
}
