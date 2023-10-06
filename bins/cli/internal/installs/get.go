package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, id string, asJSON bool) {
	view := ui.NewGetView()

	install, err := s.api.GetInstall(ctx, id)
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
