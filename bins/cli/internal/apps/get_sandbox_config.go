package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetSandboxConfig(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	sandboxCfg, err := s.api.GetAppSandboxLatestConfig(ctx, appID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(sandboxCfg)
		return nil
	}

	ui.PrintJSON(sandboxCfg)
	return nil
}
