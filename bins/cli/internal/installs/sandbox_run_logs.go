package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SandboxRunLogs(ctx context.Context, installID, runID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	log, err := s.api.GetInstallSandboxRunLogs(ctx, installID, runID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(log)
		return nil
	}

	ui.PrintLogsFromInterface(log)
	return nil
}
