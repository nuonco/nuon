package releases

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/pterm/pterm"
)

func (s *Service) Create(ctx context.Context, compID, buildID, delay string, installsPerStep int64) {
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
		if err != nil {
			releaseSpinner.Fail(err.Error() + "\n")
		}

		if release.Status == "active" {
			releaseSpinner.Success(fmt.Sprintf("successfully released component %s", release.ID))
			return
		} else {
			releaseSpinner.UpdateText(fmt.Sprintf("%s component release", release.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
