package builds

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

const (
	statusError  = "error"
	statusActive = "active"
)

func (s *Service) Create(ctx context.Context, appID, compID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	if asJSON {
		newBuild, err := s.api.CreateComponentBuild(
			ctx,
			compID,
			&models.ServiceCreateComponentBuildRequest{
				UseLatest: true,
			},
		)
		if err != nil {
			ui.PrintJSONError(err)
			return
		}

		ui.PrintJSON(newBuild)
		return
	}

	view := ui.NewCreateView("build", asJSON)
	view.Start()
	view.Update("starting component build")
	newBuild, err := s.api.CreateComponentBuild(
		ctx,
		compID,
		&models.ServiceCreateComponentBuildRequest{
			UseLatest: true,
		},
	)
	if err != nil {
		view.Fail(err)
		return
	}

	for {
		build, err := s.api.GetComponentBuild(ctx, compID, newBuild.ID)
		switch {
		case err != nil:
			view.Fail(err)
		case build.Status == statusError:
			view.Fail(fmt.Errorf("failed to create component build: %s", build.StatusDescription))
			return
		case build.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created component build %s", build.ID))
			return
		default:
			view.Update(fmt.Sprintf("%s component build", build.Status))
		}
		time.Sleep(5 * time.Second)
	}
}
