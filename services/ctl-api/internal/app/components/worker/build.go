package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Build(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusPlanning, "creating build plan")

	logStream, err := activities.AwaitCreateLogStreamByBuildID(ctx, sreq.BuildID)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("executing build")
	currentApp, err := activities.AwaitGetComponentAppByComponentID(ctx, sreq.ID)
	if err != nil {
		w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusError, "unable to get component app")
		return fmt.Errorf("unable to get component app: %w", err)
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, sreq.ID)
	if err != nil {
		w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	build, err := activities.AwaitGetComponentBuildByID(ctx, sreq.BuildID)
	if err != nil {
		w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusError, "unable to get component build")
		return fmt.Errorf("unable to get component build: %w", err)
	}

	notify := func(err error) error {
		w.sendNotification(ctx, notifications.NotificationsTypeComponentBuildFailed, currentApp.ID, map[string]string{
			"component_name": comp.Name,
			"app_name":       currentApp.Name,
			"created_by":     build.CreatedBy.Email,
		})
		return err
	}

	if comp.Status != app.ComponentStatusActive {
		w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusError, "component is not active")
		return notify(fmt.Errorf("component is not active"))
	}

	if err := w.execBuild(ctx, sreq.ID, sreq.BuildID, currentApp, sreq.SandboxMode); err != nil {
		return notify(err)
	}

	return nil
}
