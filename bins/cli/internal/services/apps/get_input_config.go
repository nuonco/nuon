package apps

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) GetInputConfig(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return err
	}

	inputCfg, err := s.api.GetAppInputLatestConfig(ctx, appID)
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(inputCfg)
		return nil
	}

	ui.PrintIndentedJSON(inputCfg)
	return nil
}
