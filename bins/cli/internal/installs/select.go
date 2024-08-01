package installs

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/pterm/pterm"
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
				installs, err = s.api.GetAllInstalls(ctx)

			} else {
				installs, err = s.api.GetAppInstalls(ctx, appID)
			}

		} else {
			installs, err = s.api.GetAllInstalls(ctx)
		}
		if err != nil {
			return view.Error(err)
		}

		if len(installs) == 0 {
			s.printNoInstallsMsg()
			return nil
		}

		// select options
		var options []string
		for _, install := range installs {
			options = append(options, fmt.Sprintf("%s: %s", install.Name, install.ID))
		}

		// select install prompt
		selectedInstall, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()
		install := strings.Split(selectedInstall, ":")

		if err := s.setInstallInConfig(ctx, strings.ReplaceAll(install[1], " ", "")); err != nil {
			return view.Error(err)
		}

		s.printInstallSetMsg(install[0], install[1])
	}

	return nil
}
