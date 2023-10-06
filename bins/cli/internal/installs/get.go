package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, installID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}
	view := ui.NewGetView()

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(install)
		return
	}

	view.Render([][]string{
		{"id", install.ID},
		{"name", install.Name},
		{"created at", install.CreatedAt},
		{"updated at", install.UpdatedAt},
		{"created by", install.CreatedByID},
		{"status", install.StatusDescription},
		{"region", install.AwsAccount.Region},
		{"role", install.AwsAccount.IamRoleArn},
	})
}
