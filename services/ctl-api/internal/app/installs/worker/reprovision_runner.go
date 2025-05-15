package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ReprovisionRunner(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID

	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: installID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	w.evClient.Send(ctx, install.RunnerID, &runnersignals.Signal{
		Type: runnersignals.OperationReprovisionServiceAccount,
	})

	// NOTE(jm): this does not send a signal at the moment
	return nil
	if err := w.evClient.SendAndWait(ctx, install.RunnerID, &runnersignals.Signal{
		Type: runnersignals.OperationReprovisionServiceAccount,
	}); err != nil {
		return errors.Wrap(err, "unable to provision service account")
	}

	return nil
}
