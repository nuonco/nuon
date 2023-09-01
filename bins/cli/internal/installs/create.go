package installs

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Create(ctx context.Context, appID, name, region, arn string) error {
	install, err := s.api.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
		Name: &name,
		AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
			Region:     region,
			IamRoleArn: &arn,
		},
	})
	if err != nil {
		return err
	}

	ui.Line(ctx, "Created new install: %s - %s", install.ID, install.Name)
	return nil
}
