package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetSandboxConfig(ctx context.Context, appID string, asJSON bool) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	sandboxCfg, err := s.api.GetAppSandboxLatestConfig(ctx, appID)
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
