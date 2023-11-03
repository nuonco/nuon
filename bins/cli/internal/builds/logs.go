package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Logs(ctx context.Context, compID, buildID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, compID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	wpLog, err := s.api.GetComponentBuildLogs(ctx, compID, buildID)
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
