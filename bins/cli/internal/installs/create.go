package installs

import (
	"context"
	"fmt"
	"strings"
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

func (s *Service) Create(ctx context.Context, appID, name, region, arn string, inputs []string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	inputsMap := make(map[string]string)
	for _, kv := range inputs {
		kvT := strings.Split(kv, "=")
		inputsMap[kvT[0]] = kvT[1]
	}

	if asJSON {
		install, err := s.api.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
			Name: &name,
			AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
				Region:     region,
				IamRoleArn: &arn,
			},
			Inputs: inputsMap,
		})
		if err != nil {
			return ui.PrintJSONError(err)
		}
		ui.PrintJSON(install)
		return nil
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
		Inputs: inputsMap,
	})
	if err != nil {
		return view.Fail(err)
	}

	for {
		ins, err := s.api.GetInstall(ctx, install.ID)
		switch {
		case err != nil:
			view.Fail(err)
		case ins.Status == statusAccessError:
			return view.Fail(fmt.Errorf("failed to create install due to access error: %s", ins.StatusDescription))
		case ins.Status == statusError:
			return view.Fail(fmt.Errorf("failed to create install: %s", ins.StatusDescription))
		case ins.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created install %s", ins.ID))
			return nil
		default:
			view.Update(fmt.Sprintf("%s install", ins.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
