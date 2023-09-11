package builds

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/pterm/pterm"
)

func (s *Service) Create(ctx context.Context, compID string) {
	buildSpinner, _ := pterm.DefaultSpinner.Start("starting component build")
	newBuild, err := s.api.CreateComponentBuild(
		ctx,
		compID,
		&models.ServiceCreateComponentBuildRequest{
			UseLatest: true,
		},
	)
	if err != nil {
		buildSpinner.Fail(err.Error() + "\n")
		return
	}

	for {
		build, err := s.api.GetComponentBuild(ctx, compID, newBuild.ID)
		if err != nil {
			buildSpinner.Fail(err.Error() + "\n")
		}

		if build.Status == "active" {
			buildSpinner.Success(fmt.Sprintf("successfully created component build %s", build.ID))
			return
		} else {
			buildSpinner.UpdateText(fmt.Sprintf("%s component build", build.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
