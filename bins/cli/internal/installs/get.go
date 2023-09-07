package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, id string) {
	view := ui.NewGetView()

	install, err := s.api.GetInstall(ctx, id)
	if err != nil {
		view.Error(err)
		return
	}

	view.Render([][]string{
		[]string{"id", install.ID},
		[]string{"name", install.Name},
		[]string{"created at", install.CreatedAt},
		[]string{"updated at", install.UpdatedAt},
		[]string{"created by", install.CreatedByID},
		[]string{"status", install.StatusDescription},
		[]string{"region", install.AwsAccount.Region},
		[]string{"role", install.AwsAccount.IamRoleArn},
	})
}
