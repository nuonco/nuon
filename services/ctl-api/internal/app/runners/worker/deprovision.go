package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Deprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.RunnerStatusDeprovisioning, "deprovisioning organization resources")

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to get runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		return w.executeDeprovisionOrgRunner(ctx, sreq.ID, sreq.SandboxMode)
	case app.RunnerGroupTypeInstall:
		return w.executeDeprovisionInstallRunner(ctx, sreq.ID, sreq.SandboxMode)
	}

	return nil
}
