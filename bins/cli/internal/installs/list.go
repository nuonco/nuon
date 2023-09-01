package installs

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) List(ctx context.Context, appID string) error {
	installs := []*models.AppInstall{}
	err := error(nil)
	if appID == "" {
		installs, err = s.api.GetAllInstalls(ctx)
	} else {
		installs, err = s.api.GetAppInstalls(ctx, appID)
	}
	if err != nil {
		return err
	}

	if len(installs) == 0 {
		ui.Line(ctx, "No installs of this app found")
	} else {
		for _, install := range installs {
			statusColor := ui.GetStatusColor(install.Status)
			ui.Line(ctx, "%s%s %s- %s - %s", statusColor, install.Status, ui.ColorReset, install.ID, install.Name)
		}
	}

	return nil
}
