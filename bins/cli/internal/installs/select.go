package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
)

func (s *Service) Select(ctx context.Context, appID, installID string, asJSON bool) error {
	view := ui.NewGetView()

	if installID != "" {
		s.SetCurrent(ctx, installID, asJSON)
	} else {

		var (
			installs []*models.AppInstall
			err      error
		)

		if appID != "" {
			appID, err := lookup.AppID(ctx, s.api, appID)
			if err != nil {
				installs, _, err = s.listInstalls(ctx, 0, 50)
			} else {
				installs, _, err = s.listAppInstalls(ctx, appID, 0, 50)
			}

		} else {
			installs, _, err = s.listInstalls(ctx, 0, 50)
		}
		if err != nil {
			return view.Error(err)
		}

		if len(installs) == 0 {
			s.printNoInstallsMsg()
			return nil
		}

		// Convert installs to selector options
		installOptions := make([]bubbles.InstallOption, len(installs))
		for i, install := range installs {
			installOptions[i] = bubbles.InstallOption{
				ID:   install.ID,
				Name: install.Name,
			}
		}

		// Show install selector
		selectedInstallID, err := bubbles.SelectInstall(installOptions)
		if err != nil {
			return view.Error(err)
		}

		if err := s.setInstallID(ctx, selectedInstallID); err != nil {
			return view.Error(err)
		}

		// Find selected install for display
		var selectedInstall *models.AppInstall
		for _, install := range installs {
			if install.ID == selectedInstallID {
				selectedInstall = install
				break
			}
		}

		if selectedInstall != nil {
			s.printInstallSetMsg(selectedInstall.Name, selectedInstall.ID)
		}
	}

	return nil
}
