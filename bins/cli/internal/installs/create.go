package installs

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

const (
	statusError       = "error"
	statusActive      = "active"
	statusAccessError = "access-error"
)

func (s *Service) Create(ctx context.Context, appID, name, region, arn string, asJSON bool) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	if asJSON {
		install, err := s.api.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
			Name: &name,
			AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
				Region:     region,
				IamRoleArn: &arn,
			},
		})
		if err != nil {
			ui.PrintJSONError(err)
			return
		}
		ui.PrintJSON(install)
		return
	}

	view := ui.NewCreateView("install", asJSON)
	view.Start()
	view.Update("creating install")
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

	for {
		ins, err := s.api.GetInstall(ctx, install.ID)
		switch {
		case err != nil:
			view.Fail(err)
		case ins.Status == statusAccessError:
			view.Fail(fmt.Errorf("failed to create install due to access error: %s", ins.StatusDescription))
			return
		case ins.Status == statusError:
			view.Fail(fmt.Errorf("failed to create install: %s", ins.StatusDescription))
			return
		case ins.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created install %s", ins.ID))
			return
		default:
			view.Update(fmt.Sprintf("%s install", ins.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
