package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	runnersignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) Delete(ctx workflow.Context, sreq signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	w.updateStatus(ctx, sreq.ID, app.OrgStatusDeleting, "ensuring all apps are deleted before deprovisioning")

	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		return err
	}

	err = w.Deprovision(ctx, sreq)
	if err != nil {
		if !sreq.ForceDelete {
			return err
		}

		l.Error("unable to deprovision org, continuing anyway", zap.Error(err))
	}

	runner := org.RunnerGroup.ActiveRunner()
	if runner == nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "no runners available in runner group")
		return fmt.Errorf("no runners available in runner group")
	}

	w.ev.Send(ctx, runner.ID, &runnersignals.Signal{
		Type: runnersignals.OperationDelete,
	})
	err = w.pollRunnerNotFound(ctx, sreq.ID)
	if err != nil {
		if !sreq.ForceDelete {
			return errors.Wrap(err, "unable to poll runner to not found")
		}

		l.Error("unable to poll runner to not found, continuing anyway", zap.Error(err))
	}

	// update status with response
	if err := activities.AwaitDeleteByOrgID(ctx, sreq.ID); err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to delete organization from database")
		return fmt.Errorf("unable to delete org: %w", err)
	}
	return nil
}
