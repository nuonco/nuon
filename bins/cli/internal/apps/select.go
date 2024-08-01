package apps

import (
	"context"
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/pterm/pterm"
)

func (s *Service) Select(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewGetView()

	if appID != "" {
		return s.SetCurrent(ctx, appID, asJSON)
	} else {
		apps, err := s.api.GetApps(ctx)
		if err != nil {
			return view.Error(err)
		}

		if len(apps) == 0 {
			s.printNoAppsMsg()
			return nil
		}

		// select options
		var options []string
		for _, app := range apps {
			options = append(options, fmt.Sprintf("%s: %s", app.Name, app.ID))
		}

		// select app prompt
		selectedApp, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()
		app := strings.Split(selectedApp, ":")

		if err := s.setAppInConfig(ctx, strings.ReplaceAll(app[1], " ", "")); err != nil {
			return view.Error(err)
		}

		s.printAppSetMsg(app[0], app[1])
	}
	return nil
}
