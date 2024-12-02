package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Delete(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := helpers.AwaitGetInstallByID(ctx, sreq.ID)
	if err != nil {
		return err
	}
	installID := sreq.ID

	// fail all queued deploys
	if err := activities.AwaitFailQueuedDeploysByInstallID(ctx, installID); err != nil {
		return fmt.Errorf("unable to fail queued install: %w", err)
	}

	if err := w.Deprovision(ctx, sreq); err != nil {
		return err
	}

	w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationDelete,
	})
	if err := w.pollRunnerNotFound(ctx, install.RunnerGroup.Runners[0].ID); err != nil {
		return err
	}

	// update status with response
	if err := activities.AwaitDeleteByInstallID(ctx, installID); err != nil {
		return fmt.Errorf("unable to delete install: %w", err)
	}

	return nil
}
