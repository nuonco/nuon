package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) InstallStackVersionRun(ctx workflow.Context, sreq signals.RequestSignal) error {
	runner, err := activities.AwaitGetRunnerByID(ctx, sreq.ID)
	if err != nil {
		return err
	}

	if runner.Status != app.RunnerStatusAwaitingInstallStackRun {
		return nil
	}

	w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "runner install stack was run, waiting for health check to mark healthy")
	return nil
}
