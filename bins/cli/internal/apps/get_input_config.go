package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetInputConfig(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	inputCfg, err := s.api.GetAppInputLatestConfig(ctx, appID)
	if err != nil {
		return view.Error(err)
	}

	// NOTE: ignore json flag and always output json
	ui.PrintJSON(inputCfg)

	return nil
}
