package builds

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

func (s *Service) Create(ctx context.Context, compID string, asJSON bool) {
	if asJSON == true {
		newBuild, err := s.api.CreateComponentBuild(
			ctx,
			compID,
			&models.ServiceCreateComponentBuildRequest{
				UseLatest: true,
			},
		)
		if err != nil {
			fmt.Printf("failed to create build: %s", err)
			return
		}
		j, _ := json.Marshal(newBuild)
		fmt.Println(string(j))
	} else {
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
			switch {
			case err != nil:
				buildSpinner.Fail(err.Error() + "\n")
			case build.Status == statusError:
				buildSpinner.Fail(fmt.Errorf("failed to create component build: %s", build.StatusDescription))
				return
			case build.Status == statusActive:
				buildSpinner.Success(fmt.Sprintf("successfully created component build %s", build.ID))
				return
			default:
				buildSpinner.UpdateText(fmt.Sprintf("%s component build", build.Status))
			}
			time.Sleep(5 * time.Second)
		}
	}
}
