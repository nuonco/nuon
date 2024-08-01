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

func (s *Service) Create(ctx context.Context, appID, compID, buildID, delay string, autoBuild, latestBuild bool, installsPerStep int64, asJSON bool) error {
	var err error
	view := ui.NewCreateView("release", asJSON)
	view.Start()

	if buildID == "" && compID == "" {
		return view.Fail(errMissingInput)
	}
	if !latestBuild && !autoBuild && buildID == "" {
		return view.Fail(errMissingBuildInput)
	}

	req := &models.ServiceCreateComponentReleaseRequest{
		BuildID: buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			Delay:           delay,
			InstallsPerStep: installsPerStep,
		},
	}

	if compID != "" {
		compID, err = lookup.ComponentID(ctx, s.api, appID, compID)
		if err != nil {
			return ui.PrintError(err)
		}
		view.Update(fmt.Sprintf("using component %s", compID))
	}

	// if we weren't given a build ID, we get the latest build for the component
	if latestBuild {
		view.Update(fmt.Sprintf("getting latest build for component %s", compID))
		var latestBuild *models.AppComponentBuild
		latestBuild, err = s.api.GetComponentLatestBuild(ctx, compID)
		if err != nil {
			return view.Fail(err)
		}

		req.BuildID = latestBuild.ID
		view.Update(fmt.Sprintf("using latest build %s", buildID))
	} else if autoBuild {
		req.AutoBuild = true
		view.Update("automatically triggering a new build")
	} else {
		compBuild, err := s.api.GetBuild(ctx, buildID)
		if err != nil {
			return view.Fail(err)
		}
		compID = compBuild.ComponentID
		view.Update(fmt.Sprintf("using component %s", compID))
	}

	if compID == "" {
		return view.Fail(fmt.Errorf("no component id was able to be found"))
	}

	view.Update("creating release")
	release, err := s.api.CreateComponentRelease(ctx, compID, req)

	// Need to refactor to move this view logic into view.
	if asJSON {
		if err != nil {
			return ui.PrintJSONError(err)
		}
		ui.PrintJSON(release)
		return nil
	}

	if err != nil {
		return view.Fail(err)
	}

	for {
		release, err := s.api.GetRelease(ctx, release.ID)

		switch {
		case err != nil:
			return view.Fail(err)
		case release.Status == statusError:
			return view.Fail(fmt.Errorf(release.StatusDescription))
		case release.Status == statusActive:
			view.Success(release.ID)
			return nil
		default:
			view.Update(fmt.Sprintf("release %s %s", release.ID, release.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
