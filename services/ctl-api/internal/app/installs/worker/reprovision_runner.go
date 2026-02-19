package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
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

	for _, runner := range install.RunnerGroup.Runners {
		if runner.Platform == app.AppRunnerTypeLocal {
			continue
		}
		w.evClient.Send(ctx, runner.ID, &runnersignals.Signal{
			Type: runnersignals.OperationReprovisionServiceAccount,
		})
	}

	return nil
}
