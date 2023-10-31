package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) PrintDeployPlan(ctx context.Context, installID string, deployID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	plan, err := s.api.GetInstallDeployPlan(ctx, installID, deployID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(plan)
		return
	}

	ui.PrintJSON(plan)
}
