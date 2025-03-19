package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

type ShutdownJobRequest struct {
	RunnerID      string `validate:"required"`
	HealthCheckID string `validate:"required"`
}

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Shutdown(ctx workflow.Context, sreq signals.RequestSignal) error {
	runner, err := activities.AwaitGetByRunnerID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	// create a logstream derived from the healthcheck
	logStream, err := activities.AwaitCreateLogStreamByOperationID(ctx, fmt.Sprintf("%s-ttl-shutdown", sreq.HealthCheckID))
	if err != nil {
		return errors.Wrap(err, "unable to create log stream for health check")
	}
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "could not get logger")
	}

	runnerJob, err := activities.AwaitCreateShutdownJob(ctx, &activities.CreateShutdownJobRequest{
		RunnerID:    runner.ID,
		OwnerID:     sreq.HealthCheckID,
		LogStreamID: logStream.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to create shutdown job")
		return errors.Wrap(err, "unable to create job")
	}

	l.Info("dispatching shutdown job to runner",
		zap.String("runner_id", runner.Org.RunnerGroup.Runners[0].ID),
		zap.String("runner_type", string(runner.RunnerGroup.Type)),
	)

	// We have to send the signal and then return to allow it to be processed.
	// Waiting for it to complete would deadlock. Not a big deal because
	// we wouldn't do anything differently even if it failed.
	w.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type:  signals.OperationProcessJob,
		JobID: runnerJob.ID,
	})

	return nil
}
