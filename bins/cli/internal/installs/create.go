package installs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	statusError       = "error"
	statusActive      = "active"
	statusAccessError = "access_error"
)

func (s *Service) Create(ctx context.Context, appID, name, region, arn string, inputs []string, asJSON, noSelect bool) error {
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
		case ins.SandboxStatus == statusAccessError:
			return view.Fail(fmt.Errorf("failed to create install due to access error: %s", ins.StatusDescription))
		case ins.SandboxStatus == statusError:
			return view.Fail(fmt.Errorf("failed to create install: %s", ins.StatusDescription))
		case ins.SandboxStatus == statusActive:
			view.Success(fmt.Sprintf("successfully created install %s", ins.ID))
			if !noSelect {
				if err := s.setInstallID(ctx, ins.ID); err == nil {
					s.printInstallSetMsg(name, ins.ID)
				} else {
					view.Fail(errs.NewUserFacing("failed to set install as current: %s", err))
				}
				return nil
			}
			return nil
		default:
			view.Update(fmt.Sprintf("%s install", ins.SandboxStatus))
		}

		time.Sleep(5 * time.Second)
	}
}
