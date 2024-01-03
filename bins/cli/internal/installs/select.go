package installs

import (
	"context"
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/pterm/pterm"
)

func (s *Service) Select(ctx context.Context, appID, installID string, asJSON bool) {
	view := ui.NewGetView()

	if installID != "" {
		s.SetCurrent(ctx, installID, asJSON)
	} else {

		installs, err := s.api.GetAppInstalls(ctx, appID)
		if err != nil {
			view.Error(err)
			return
		}

		if len(installs) == 0 {
			s.printNoInstallsMsg()
			return
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
			view.Error(err)
			return
		}

		s.printInstallSetMsg(install[0], install[1])
	}
}
