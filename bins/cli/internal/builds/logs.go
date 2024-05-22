package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Logs(ctx context.Context, appID, compID, buildID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	log, err := s.api.GetComponentBuildLogs(ctx, compID, buildID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(log)
		return
	}

	ui.PrintBuildLog(log)
}
