package releases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/pterm/pterm"
)

const (
	statusError  = "error"
	statusActive = "active"
)

func (s *Service) Create(ctx context.Context, compID, buildID, delay string, installsPerStep int64, asJSON bool) {
	if asJSON == true {
		newRelease, err := s.api.CreateComponentRelease(ctx, compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				Delay:           delay,
				InstallsPerStep: installsPerStep,
			},
		})
		if err != nil {
			fmt.Printf("failed to create release: %s", err)
			return
		}
		j, _ := json.Marshal(newRelease)
		fmt.Println(string(j))
	} else {
		releaseSpinner, _ := pterm.DefaultSpinner.Start("starting release")

		newRelease, err := s.api.CreateComponentRelease(ctx, compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				Delay:           delay,
				InstallsPerStep: installsPerStep,
			},
		})
		if err != nil {
			releaseSpinner.Fail(err.Error() + "\n")
			return
		}

		for {
			release, err := s.api.GetRelease(ctx, newRelease.ID)
			switch {
			case err != nil:
				releaseSpinner.Fail(err.Error() + "\n")
			case release.Status == statusError:
				releaseSpinner.Fail(fmt.Errorf("failed to create release: %s", release.StatusDescription))
				return
			case release.Status == statusActive:
				releaseSpinner.Success(fmt.Sprintf("successfully created release %s", release.ID))
				return
			default:
				releaseSpinner.UpdateText(fmt.Sprintf("%s component release", release.Status))
			}

			time.Sleep(5 * time.Second)
		}
	}
}
