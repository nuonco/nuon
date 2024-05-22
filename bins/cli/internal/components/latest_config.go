package components

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) LatestConfig(ctx context.Context, appID, compID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	config, err := s.api.GetComponentLatestConfig(ctx, compID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(config)
		return
	}

	ui.PrintJSON(config)
}
