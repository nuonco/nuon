package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetInputConfig(ctx context.Context, appID string, asJSON bool) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	sandboxCfg, err := s.api.GetAppInputLatestConfig(ctx, appID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(sandboxCfg)
		return
	}

	ui.PrintJSON(sandboxCfg)
}
