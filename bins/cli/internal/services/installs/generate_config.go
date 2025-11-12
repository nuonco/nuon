package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GenerateConfig(ctx context.Context, installID string) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	installCfgBytes, err := s.api.GenerateCLIInstallConfig(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	ui.PrintRaw(string(installCfgBytes))

	return nil
}
