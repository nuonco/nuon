package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) Select(ctx context.Context, appID, installID string, asJSON bool) error {
	view := ui.NewGetView()

	if installID != "" {
		s.SetCurrent(ctx, installID, asJSON)
		return nil
	}

	var (
		installs    []*models.AppInstall
		err         error
		resolvedApp *models.AppApp
	)

	if appID != "" {
		resolvedAppID, lookupErr := lookup.AppID(ctx, s.api, appID)
		if lookupErr != nil {
			// app lookup failed, fall back to listing all installs
			installs, _, err = s.listInstalls(ctx, 0, 50)
		} else {
			app, appErr := s.api.GetApp(ctx, resolvedAppID)
			if appErr == nil {
				resolvedApp = app
			}
			installs, _, err = s.listAppInstalls(ctx, resolvedAppID, 0, 50)
		}
	} else {
		installs, _, err = s.listInstalls(ctx, 0, 50)
	}
	if err != nil {
		return view.Error(err)
	}

	if len(installs) == 0 {
		if resolvedApp != nil {
			ui.PrintLn(fmt.Sprintf(
				"The selected app (%s), has no installs.\n nuon installs create to create one\n nuon apps deselect   to select from all installs",
				resolvedApp.Name,
			))
		} else {
			s.printNoInstallsMsg()
		}
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
	selectedInstallID, err := bubbles.SelectInstall(installOptions, s.cfg.Interactive)
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

	if selectedInstall == nil {
		return nil
	}

	s.printInstallSetMsg(selectedInstall.Name, selectedInstall.ID)
	// if the app is not set, go ahead and set it as well
	selectedAppID := selectedInstall.AppID
	if s.cfg.AppID != selectedAppID {
		err := s.setAppID(ctx, selectedAppID)
		if err != nil {
			return view.Error(err)
		}
		s.printAppSetMsg(selectedAppID)
	}

	return nil
}
