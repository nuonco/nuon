package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeleteComponent(ctx context.Context, installID, componentID string, force, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	_, err = s.api.DeleteInstallComponent(ctx, installID, componentID, force)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	ui.PrintLn("successfully scheduled teardown of install component")
	return nil
}
