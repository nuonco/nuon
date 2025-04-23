package worker

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/poll"
)

// @temporal-gen workflow
// @execution-timeout 1h
// @task-timeout 30s
func (w *Workflows) AwaitRunnerHealthy(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "no install found")
	}

	if err := poll.Poll(ctx, w.v, poll.PollOpts{
		MaxTS:           workflow.Now(ctx).Add(time.Hour),
		InitialInterval: time.Second * 15,
		MaxInterval:     time.Minute * 1,
		BackoffFactor:   1.1,
		Fn: func(ctx workflow.Context) error {
			runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
			if err != nil {
				return err
			}

			if runner.Status != app.RunnerStatusActive {
				return errors.New("runner was not healthy")
			}
			return nil
		},
	}); err != nil {
		return errors.Wrap(err, "runner was not healthy")
	}

	return nil
}
