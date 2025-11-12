package components

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) LatestConfig(ctx context.Context, appID, compID string, asJSON bool) error {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	config, err := s.api.GetComponentLatestConfig(ctx, compID)
	if err != nil {
		return view.Error(err)
	}

	ui.PrintJSON(config)
	return nil
}
