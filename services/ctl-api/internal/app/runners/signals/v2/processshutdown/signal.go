package processshutdown

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "process_shutdown"

type Signal struct {
	RunnerID     string `json:"runner_id"`
	ProcessID    string `json:"process_id"`
	ShutdownType string `json:"shutdown_type"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.ProcessID == "" {
		return errors.New("process_id is required")
	}

	_, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return errors.Wrap(err, "runner process not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get process and find the requested shutdown
	process, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner process")
	}

	// Find the most recent requested shutdown
	var shutdownID string
	for _, shutdown := range process.Shutdowns {
		if shutdown.Status == app.RunnerProcessShutdownStatusRequested {
			shutdownID = shutdown.ID
			break
		}
	}

	if shutdownID == "" {
		return nil
	}

	// Update shutdown status to in-progress
	_, err = activities.AwaitUpdateRunnerProcessShutdownStatusByShutdownID(ctx, shutdownID,
		app.RunnerProcessShutdownStatusInProgress, "shutdown in progress")
	if err != nil {
		return errors.Wrap(err, "unable to update shutdown status to in-progress")
	}

	// Wait for the runner process to report shut-down status
	// Poll with timeout
	timeout := workflow.NewTimer(ctx, 5*time.Minute)
	pollInterval := workflow.NewTimer(ctx, 10*time.Second)

	for {
		selector := workflow.NewSelector(ctx)

		timedOut := false
		selector.AddFuture(timeout, func(f workflow.Future) {
			timedOut = true
		})

		selector.AddFuture(pollInterval, func(f workflow.Future) {})

		selector.Select(ctx)

		if timedOut {
			// Timeout: mark shutdown as failed
			_, _ = activities.AwaitUpdateRunnerProcessShutdownStatusByShutdownID(ctx, shutdownID,
				app.RunnerProcessShutdownStatusFailed, "shutdown timed out waiting for process to stop")
			return errors.New("shutdown timed out")
		}

		// Check process status
		updated, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
		if err != nil {
			return errors.Wrap(err, "unable to poll runner process status")
		}

		if updated.Status == app.RunnerProcessStatusShutDown {
			// Process confirmed shut down
			_, err = activities.AwaitUpdateRunnerProcessShutdownStatusByShutdownID(ctx, shutdownID,
				app.RunnerProcessShutdownStatusCompleted, "shutdown completed")
			if err != nil {
				return errors.Wrap(err, "unable to update shutdown status to completed")
			}
			return nil
		}

		// Reset poll timer
		pollInterval = workflow.NewTimer(ctx, 10*time.Second)
	}
}
