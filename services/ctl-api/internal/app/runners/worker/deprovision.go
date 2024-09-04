package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deprovision(ctx workflow.Context, runnerID string, sandboxMode bool) error {
	w.updateStatus(ctx, runnerID, app.RunnerStatusDeprovisioning, "deprovisioning organization resources")

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to get runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		return w.executeDeprovisionOrgRunner(ctx, runnerID, sandboxMode)
	case app.RunnerGroupTypeInstall:
		return w.executeDeprovisionInstallRunner(ctx, runnerID, sandboxMode)
	}

	return nil
}
