package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

func (w *Workflows) pollDependencies(ctx workflow.Context, releaseID string) error {
	for {
		var release app.ComponentRelease
		if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
			ReleaseID: releaseID,
		}, &release); err != nil {
			w.updateStatus(ctx, releaseID, app.ReleaseStatusError, "unable to get release from database")
			return fmt.Errorf("unable to get release: %w", err)
		}

		if release.ComponentBuild.Status == "active" {
			return nil
		}
		if release.ComponentBuild.Status == "error" {
			w.updateStatus(ctx, releaseID, app.ReleaseStatusError, "build failed")
			return fmt.Errorf("build failed: %s", release.StatusDescription)
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
