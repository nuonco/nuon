package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Deprovision(ctx context.Context, installID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	err = s.api.DeprovisionInstall(ctx, installID)
	if err != nil {
		ui.PrintJSONError(err)
		return
	}

	ui.PrintLn("successfully triggered install deprovision")
}
