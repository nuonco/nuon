package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
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

	for _, runner := range install.RunnerGroup.Runners {
		if runner.Platform == app.AppRunnerTypeLocal {
			continue
		}
		w.evClient.Send(ctx, runner.ID, &runnersignals.Signal{
			Type: runnersignals.OperationProvisionServiceAccount,
		})
	}

	return nil
}
