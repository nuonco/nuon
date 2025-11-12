package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Reprovision(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	err = s.api.ReprovisionInstall(ctx, installID)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	ui.PrintLn("successfully triggered install reprovision")
	return nil
}
