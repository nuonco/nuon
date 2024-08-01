package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeployLogs(ctx context.Context, installID, deployID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	log, err := s.api.GetInstallDeployLogs(ctx, installID, deployID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(log)
		return nil
	}

	ui.PrintDeployLogs(log)
	return nil
}
