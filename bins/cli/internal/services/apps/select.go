package apps

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
)

func (s *Service) Select(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewGetView()

	if appID != "" {
		return s.SetCurrent(ctx, appID, asJSON)
	} else {
		apps, _, err := s.listApps(ctx, 0, 50)
		if err != nil {
			return view.Error(err)
		}

		if len(apps) == 0 {
			s.printNoAppsMsg()
			return nil
		}

		// Convert apps to selector options
		appOptions := make([]bubbles.AppOption, len(apps))
		for i, app := range apps {
			appOptions[i] = bubbles.AppOption{
				ID:   app.ID,
				Name: app.Name,
			}
		}

		// Show app selector
		selectedAppID, err := bubbles.SelectApp(appOptions)
		if err != nil {
			return view.Error(err)
		}

		if err := s.setAppID(ctx, selectedAppID); err != nil {
			return view.Error(err)
		}

		// Find selected app for display
		var selectedApp *models.AppApp
		for _, app := range apps {
			if app.ID == selectedAppID {
				selectedApp = app
				break
			}
		}

		if selectedApp != nil {
			s.printAppSetMsg(selectedApp.Name, selectedApp.ID)
		}
	}
	return nil
}
