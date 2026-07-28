package components

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) LatestConfig(ctx context.Context, appID, compID string, asJSON bool) error {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	config, err := s.api.GetComponentLatestConfig(ctx, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		ui.PrintJSON(config)
		return nil
	}

	ui.PrintIndentedJSON(config)
	return nil
}
