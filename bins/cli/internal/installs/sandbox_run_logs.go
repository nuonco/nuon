package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SandboxRunLogs(ctx context.Context, installID, runID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	log, err := s.api.GetInstallSandboxRunLogs(ctx, installID, runID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(log)
		return
	}

	ui.PrintLogsFromInterface(log)
}
