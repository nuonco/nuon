package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ProvisionRunner(ctx workflow.Context, sreq signals.RequestSignal) error {
	// NOTE(jm): this is only used because we do not have a good way to send a signal to another namespace via a
	// step
	installID := sreq.ID

	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: installID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	w.evClient.Send(ctx, install.RunnerID, &runnersignals.Signal{
		Type: runnersignals.OperationProvisionServiceAccount,
	})

	return nil
}
