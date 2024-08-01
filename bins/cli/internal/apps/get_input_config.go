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

	if asJSON {
		ui.PrintJSON(inputCfg)
		// TODO (sdboyer) this seems like a bug, should it always print JSON?
		// } else {
		// 	ui.PrintJSON(inputCfg)
	}

	return nil
}
