package dev

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
)

func (s *Service) pollDeploys(ctx context.Context, installID string, deploys []*models.AppInstallDeploy) error {
	// Early return if no deploys to monitor
	if len(deploys) == 0 {
		return nil
	}

	depByID := make(map[string]*models.AppInstallDeploy)
	for _, dep := range deploys {
		depByID[dep.ID] = dep
	}

	pollTimeout, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	multiSpinner := bubbles.NewMultiSpinnerView()
	
	// Add all spinners first
	for _, dep := range deploys {
		multiSpinner.AddSpinner(dep.ID, fmt.Sprintf("deploying %s", dep.ComponentName))
	}
	
	// Then start the display
	multiSpinner.Start()

	time.Sleep(time.Second * 5)

	var deploysFailed error
	for {
		select {
		case <-pollTimeout.Done():
			err := fmt.Errorf("timeout waiting for components to deploy")
			ui.PrintError(err)
			for depID := range depByID {
				dep := depByID[depID]
				multiSpinner.CompleteSpinner(dep.ID, false, fmt.Sprintf("timeout waiting for %s to deploy", dep.ComponentName))
			}
			multiSpinner.Stop()
			return err
		default:
		}

		completedDeploys := make([]string, 0)
		
		for depID := range depByID {
			dep := depByID[depID]
			installDeploy, err := s.api.GetInstallDeploy(ctx, installID, dep.ID)
			if err != nil {
				if nuon.IsServerError(err) {
					multiSpinner.CompleteSpinner(dep.ID, false, fmt.Sprintf("error deploying %s", dep.ComponentName))
					completedDeploys = append(completedDeploys, depID)
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
				deploysFailed = errors.New("deploys failed")
				multiSpinner.CompleteSpinner(dep.ID, false, fmt.Sprintf("error deploying %s", dep.ComponentName))
				completedDeploys = append(completedDeploys, depID)
				continue
			}

			if installDeploy.Status == "active" {
				multiSpinner.CompleteSpinner(dep.ID, true, fmt.Sprintf("finished deploying %s", dep.ComponentName))
				completedDeploys = append(completedDeploys, depID)
				continue
			}
		}

		// Remove completed deploys from tracking
		for _, depID := range completedDeploys {
			delete(depByID, depID)
		}

		if len(depByID) == 0 {
			multiSpinner.Stop()
			return deploysFailed
		}

		time.Sleep(defaultSyncSleep)
	}
}
