package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) isBuildDeployable(bld *app.ComponentBuild) bool {
	return bld.Status == app.ComponentBuildStatusActive
}

func (w *Workflows) pollForDeployableBuild(ctx workflow.Context, installDeployId, componentBuildID string) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	bld, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, componentBuildID)
	if err != nil {
		return errors.Wrap(err, "unable to get component build")
	}

	if w.isBuildDeployable(bld) {
		l.Info("build is deployable")
		return nil
	}

	l.Info("build is not yet deployable, polling")
	sleepTimer := time.Second * 10
	maxAttempts := 20
	attempt := 0
	for {
		if attempt >= maxAttempts {
			return fmt.Errorf("build is not deployable after %d polling attempts", maxAttempts)
		}

		attempt++

		// Get the latest build
		bld, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, bld.ID)
		if err != nil {
			return fmt.Errorf("unable to get component build: %w", err)
		}

		// Check if the build is deployable
		if w.isBuildDeployable(bld) {
			return nil
		}

		if bld.Status == app.ComponentBuildStatusError {
			l.Error("component build is in an error state")
			return fmt.Errorf("component build is in an error state")
		}

		workflow.Sleep(ctx, sleepTimer)
	}
}
