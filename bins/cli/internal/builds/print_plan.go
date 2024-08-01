package builds

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) PrintPlan(ctx context.Context, appID, compID, buildID string, asJSON bool) error {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	build, err := s.api.GetComponentBuildPlan(ctx, compID, buildID)
	if err != nil {
		return view.Error(errors.Wrap(err, "failed to fetch component build plan"))
	}

	if asJSON {
		ui.PrintJSON(build)
		return nil
	}

	ui.PrintJSON(build)
	return nil
}
