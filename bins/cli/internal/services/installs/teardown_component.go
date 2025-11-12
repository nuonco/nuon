package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) TeardownComponent(ctx context.Context, installID, componentID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	err = s.api.TeardownInstallComponent(ctx, installID, componentID)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	ui.PrintLn("successfully triggered teardown")
	return nil
}
