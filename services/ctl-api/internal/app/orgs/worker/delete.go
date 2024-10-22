package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) Delete(ctx workflow.Context, sreq signals.RequestSignal) error {
	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		return err
	}

	if err := w.Deprovision(ctx, sreq); err != nil {
		return err
	}

	w.ev.Send(ctx, org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationDelete,
	})
	if err := w.pollRunnerNotFound(ctx, sreq.ID); err != nil {
		return err
	}

	// update status with response
	if err := activities.AwaitDeleteByOrgID(ctx, sreq.ID); err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to delete organization from database")
		return fmt.Errorf("unable to delete org: %w", err)
	}
	return nil
}
