package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeployLogs(ctx context.Context, installID, deployID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	wpLog, err := s.api.GetInstallDeployLogs(ctx, installID, deployID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(wpLog)
		return
	}

	ui.PrintJSON(wpLog)
}
