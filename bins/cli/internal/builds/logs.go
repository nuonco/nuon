package builds

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Logs(ctx context.Context, appID, compID, buildID string, asJSON bool) error {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	log, err := s.api.GetComponentBuildLogs(ctx, compID, buildID)
	if err != nil {
		return view.Error(errors.Wrap(err, "failed to fetch component build logs"))
	}

	if asJSON {
		ui.PrintJSON(log)
	} else {
		ui.PrintBuildLogs(log)
	}

	return nil
}
