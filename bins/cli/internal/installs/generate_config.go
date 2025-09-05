package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
)

func (s *Service) GenerateConfig(ctx context.Context, installID string) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	view := ui.NewGetView()

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	curInps, err := s.api.GetInstallCurrentInputs(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	appInputCfg, err := s.api.GetAppInputLatestConfig(ctx, install.AppID)

	var ins config.Install
	ins.ParseIntoInstall(install, curInps, appInputCfg, true)

	ui.PrintTOML(ins)

	return nil
}
