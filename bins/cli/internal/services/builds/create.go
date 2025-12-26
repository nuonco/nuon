package builds

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/errs"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	statusError  = "error"
	statusActive = "active"
)

func (s *Service) Create(ctx context.Context, appID, compID string, asJSON bool) error {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
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
		} else {
			ui.PrintJSON(newBuild)
		}

		return err
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
		return view.Fail(errors.Wrap(err, "error creating build"))
	}

	for {
		build, err := s.api.GetComponentBuild(ctx, compID, newBuild.ID)
		switch {
		case err != nil:
			view.Fail(err)
		case build.Status == statusError:
			return view.Fail(errs.NewUserFacing("component build encountered an error: %s", build.StatusDescription))
		case build.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created component build %s", build.ID))
			return nil
		default:
			view.Update(fmt.Sprintf("%s component build", build.Status))
		}
		time.Sleep(5 * time.Second)
	}
}
