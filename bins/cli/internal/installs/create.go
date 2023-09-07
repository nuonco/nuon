package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (s *Service) Create(ctx context.Context, appID, name, region, arn string) {
	view := ui.NewCreateView("install")

	view.Start()
	install, err := s.api.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
		Name: &name,
		AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
			Region:     region,
			IamRoleArn: &arn,
		},
	})
	if err != nil {
		view.Fail(err)
		return
	}

	view.Success(install.ID)
}
