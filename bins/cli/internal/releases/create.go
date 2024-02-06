package releases

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

var errMissingInput = fmt.Errorf("need either a build ID or a component ID")
var errMissingBuildInput = fmt.Errorf("must pass in one of --build-id, --auto-build or --latest-build")

func (s *Service) Create(ctx context.Context, compID, buildID, delay string, autoBuild, latestBuild bool, installsPerStep int64, asJSON bool) {
	var err error
	view := ui.NewCreateView("release", asJSON)
	view.Start()

	if buildID == "" && compID == "" {
		view.Fail(errMissingInput)
		return
	}
	if !latestBuild && !autoBuild && buildID == "" {
		view.Fail(errMissingBuildInput)
		return
	}

	req := &models.ServiceCreateComponentReleaseRequest{
		BuildID: buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			Delay:           delay,
			InstallsPerStep: installsPerStep,
		},
	}

	if compID != "" {
		compID, err = lookup.ComponentID(ctx, s.api, compID)
		if err != nil {
			ui.PrintError(err)
			return
		}
		view.Update(fmt.Sprintf("using component %s", compID))
	}

	// if we weren't given a build ID, we get the latest build for the component
	if latestBuild {
		view.Update(fmt.Sprintf("getting latest build for component %s", compID))
		var latestBuild *models.AppComponentBuild
		latestBuild, err = s.api.GetComponentLatestBuild(ctx, compID)
		if err != nil {
			view.Fail(err)
			return
		}

		req.BuildID = latestBuild.ID
		view.Update(fmt.Sprintf("using latest build %s", buildID))
	} else if autoBuild {
		req.AutoBuild = true
		view.Update("automatically triggering a new build")
	} else {
		compBuild, err := s.api.GetBuild(ctx, buildID)
		if err != nil {
			view.Fail(err)
			return
		}
		compID = compBuild.ComponentID
		view.Update(fmt.Sprintf("using component %s", compID))
	}

	if compID == "" {
		view.Fail(fmt.Errorf("no component id was able to be found"))
		return
	}

	view.Update("creating release")
	release, err := s.api.CreateComponentRelease(ctx, compID, req)

	// Need to refactor to move this view logic into view.
	if asJSON {
		if err != nil {
			ui.PrintJSONError(err)
			return
		}
		ui.PrintJSON(release)
		return
	}

	if err != nil {
		view.Fail(err)
		return
	}

	for {
		release, err := s.api.GetRelease(ctx, release.ID)

		switch {
		case err != nil:
			view.Fail(err)
			return
		case release.Status == statusError:
			view.Fail(fmt.Errorf(release.StatusDescription))
			return
		case release.Status == statusActive:
			view.Success(release.ID)
			return
		default:
			view.Update(fmt.Sprintf("release %s %s", release.ID, release.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
