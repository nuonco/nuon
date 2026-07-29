package apps

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) GetSandboxConfig(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	sandboxCfg, err := s.api.GetAppSandboxLatestConfig(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		ui.PrintJSON(sandboxCfg)
		return nil
	}

	ui.PrintIndentedJSON(sandboxCfg)
	return nil
}
