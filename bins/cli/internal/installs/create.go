package installs

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/pterm/pterm"
)

const (
	statusError  = "error"
	statusActive = "active"
)

func (s *Service) Create(ctx context.Context, appID, name, region, arn string) {
	view, _ := pterm.DefaultSpinner.Start("creating install")

	view.Start()
	install, err := s.api.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
		Name: &name,
		AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
			Region:     region,
			IamRoleArn: &arn,
		},
	})
	if err != nil {
		view.Fail(err.Error() + "\n")
		return
	}

	for {
		ins, err := s.api.GetInstall(ctx, install.ID)
		switch {
		case err != nil:
			view.Fail(err.Error() + "\n")
		case ins.Status == statusError:
			view.Fail(fmt.Errorf("failed to create install: %s", ins.StatusDescription))
			return
		case ins.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created install %s", ins.ID))
			return
		default:
			view.UpdateText(fmt.Sprintf("%s install", ins.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
