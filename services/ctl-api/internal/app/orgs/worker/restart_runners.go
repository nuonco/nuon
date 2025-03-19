package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 10s
func (w *Workflows) RestartRunners(ctx workflow.Context, sreq signals.RequestSignal) error {
	runners, err := activities.AwaitGetRunnersByID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runners")
	}

	for _, ev := range runners {
		if err := w.restartRunner(ctx, ev.ID); err != nil {
			return errors.Wrap(err, "unable to restart runner")
		}
	}

	return nil
}

func (w *Workflows) restartRunner(ctx workflow.Context, runnerID string) error {
	w.ev.Send(ctx, runnerID, &runnersignals.Signal{
		Type: runnersignals.OperationShutdown,
	})
	return nil
}
