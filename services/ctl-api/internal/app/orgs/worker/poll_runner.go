package worker

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/runnerstatussignals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	pollRunnerTimeout time.Duration = time.Minute * 5
	pollRunnerPeriod  time.Duration = time.Second * 10
)

// pollRunner blocks until the runner reaches Active status or an error status,
// using the runner_status_wakeup signal channel instead of a polling loop.
// Writers of Runner.Status send this signal so the workflow wakes immediately
// rather than waiting for the next 10-second tick.
func (w *Workflows) pollRunner(ctx workflow.Context, runnerID string) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	signalChan := workflow.GetSignalChannel(ctx, runnerstatussignals.SignalName)
	deadline := workflow.Now(ctx).Add(pollRunnerTimeout)

	for {
		runner, err := activities.AwaitGetRunnerByID(ctx, runnerID)
		if err != nil {
			return fmt.Errorf("unable to get runner from database: %w", err)
		}

		if runner.Status == app.RunnerStatusActive {
			return nil
		}
		if runner.Status == app.RunnerStatusError {
			return fmt.Errorf("runner is in error state")
		}

		sel := workflow.NewSelector(ctx)

		var wakeup runnerstatussignals.WakeUp
		sel.AddReceive(signalChan, func(c workflow.ReceiveChannel, _ bool) {
			c.Receive(ctx, &wakeup)
		})

		timedOut := false
		if d := deadline.Sub(workflow.Now(ctx)); d > 0 {
			sel.AddFuture(workflow.NewTimer(ctx, d), func(workflow.Future) {
				timedOut = true
			})
		} else {
			timedOut = true
		}

		sel.Select(ctx)

		if timedOut {
			return fmt.Errorf("runner did not become active within %s", pollRunnerTimeout)
		}

		l.Debug("runner status wakeup", zap.String("reason", wakeup.Reason))
	}
}

func (w *Workflows) pollRunnerNotFound(ctx workflow.Context, runnerID string) error {
	timeout := workflow.Now(ctx).Add(pollRunnerTimeout)

	var lastStatus app.RunnerStatus
	for !workflow.Now(ctx).After(timeout) {
		runner, err := activities.AwaitGetRunnerByID(ctx, runnerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			return fmt.Errorf("unable to get runner from database: %w", err)
		}

		if runner.Status == app.RunnerStatusActive {
			return nil
		}

		lastStatus = runner.Status
		workflow.Sleep(ctx, pollRunnerPeriod)
	}

	return fmt.Errorf("runner did not reach status after %s - last status %s", pollRunnerTimeout, lastStatus)
}
