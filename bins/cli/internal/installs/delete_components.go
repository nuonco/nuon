package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeleteComponents(ctx context.Context, installID string, force, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	_, err = s.api.DeleteInstallComponents(ctx, installID, force)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	ui.PrintLn("successfully scheduled teardown of all install components")
	return nil
}
