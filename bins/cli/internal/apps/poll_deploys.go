package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/pterm/pterm"
)

func (s *Service) pollDeploys(ctx context.Context, install *models.AppInstall, deploys []*models.AppInstallDeploy) error {
	depByID := make(map[string]*models.AppInstallDeploy)
	for _, dep := range deploys {
		depByID[dep.ID] = dep
	}

	pollTimeout, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	multi := pterm.DefaultMultiPrinter

	spinnersByDeployID := make(map[string]*pterm.SpinnerPrinter)
	for _, dep := range deploys {
		spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start(fmt.Sprintf("deploying %s to %s", dep.ComponentName, install.Name))
		spinnersByDeployID[dep.ID] = spinner
	}

	multi.Start()

	time.Sleep(time.Second * 5)

	for {
		select {
		case <-pollTimeout.Done():
			err := fmt.Errorf("timeout waiting for components to deploy")
			ui.PrintError(err)
			for depID, spinner := range spinnersByDeployID {
				dep, _ := depByID[depID]
				spinner.Fail(fmt.Sprintf("timeout waiting for %s to deploy", dep.ComponentName))
			}
			multi.Stop()
			return err
		default:
		}

		for depID := range spinnersByDeployID {
			dep, _ := depByID[depID]
			installDeploy, err := s.api.GetInstallDeploy(ctx, install.ID, dep.ID)
			if err != nil {
				if nuon.IsServerError(err) {
					spinnersByDeployID[depID].Fail(fmt.Sprintf("error deploying %s", dep.ComponentName))
					delete(spinnersByDeployID, depID)
					continue
				}
				if nuon.IsNotFound(err) {
					continue
				}
				if installDeploy == nil {
					continue
				}
			}

			if installDeploy.Status == "error" {
				spinnersByDeployID[depID].Fail(fmt.Sprintf("error deploying %s", dep.ComponentName))
				delete(spinnersByDeployID, depID)
				continue
			}

			if installDeploy.Status == "active" {
				spinnersByDeployID[depID].Success(fmt.Sprintf("finished deploying %s", dep.ComponentName))
				delete(spinnersByDeployID, depID)
				continue
			}
		}

		if len(spinnersByDeployID) == 0 {
			multi.Stop()
			return nil
		}

		time.Sleep(defaultSyncSleep)
	}
}
