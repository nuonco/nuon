package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) GenerateConfig(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	installCfgBytes, err := s.api.GenerateCLIInstallConfig(ctx, installID)
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(map[string]string{
			"install_id": installID,
			"config":     string(installCfgBytes),
		})
		return nil
	}

	ui.PrintRaw(string(installCfgBytes))

	return nil
}
