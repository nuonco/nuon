package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) InstallStackVersionRun(ctx workflow.Context, sreq signals.RequestSignal) error {
	runner, err := activities.AwaitGetByRunnerID(ctx, sreq.ID)
	if err != nil {
		return err
	}

	if !generics.SliceContains(runner.Status, []app.RunnerStatus{
	        app.RunnerStatusAwaitingInstallStackRun, 
	        app.RunnerStatusPending, 
        }){
		return nil
	}

	w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "runner install stack was run, waiting for health check to mark healthy")
	return nil
}
