package installs

import (
	"context"

	"github.com/mitchellh/go-wordwrap"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetDeploy(ctx context.Context, installID, deployID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	view := ui.NewGetView()

	installDeploy, err := s.api.GetInstallDeploy(ctx, installID, deployID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(installDeploy)
		return nil
	}

	view.Render([][]string{
		{"install id", installDeploy.InstallID},
		{"deploy id", installDeploy.ID},
		{"build id", installDeploy.BuildID},
		{"release id", installDeploy.ReleaseID},
		{"status", installDeploy.Status},
		{"description", wordwrap.WrapString(installDeploy.StatusDescription, 75)},
	})
	return nil
}
