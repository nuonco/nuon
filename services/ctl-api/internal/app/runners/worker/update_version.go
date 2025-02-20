package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) UpdateVersion(ctx workflow.Context, sreq signals.RequestSignal) error {
	runner, err := activities.AwaitGetByRunnerID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	// Create a logstream tied to the healthcheck
	logStream, err := activities.AwaitCreateLogStreamByOperationID(ctx, sreq.HealthCheckID)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream for health check")
	}
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	runnerJob, err := activities.AwaitCreateUpdateVersionJob(ctx, &activities.CreateUpdateVersionJobRequest{
		RunnerID:    runner.Org.RunnerGroup.Runners[0].ID,
		OwnerID:     sreq.HealthCheckID,
		LogStreamID: logStream.ID,
	})
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.RunnerStatusError, "unable to create shutdown job")
		return errors.Wrap(err, "unable to create job")
	}

	// Update the healthcheck with the job it caused to happen
	err = activities.AwaitSetHealthCheckRunnerJob(ctx, activities.SetHealthCheckRunnerJobRequest{
		HealthCheckID: sreq.HealthCheckID,
		RunnerJobID:   runnerJob.ID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to set runner job on health check")
	}

	// We have to send the signal and then return to allow it to be processed.
	// Waiting for it to complete would deadlock. Not a big deal because
	// we wouldn't do anything differently even if it failed.
	w.evClient.Send(ctx, runner.Org.RunnerGroup.Runners[0].ID, &signals.Signal{
		Type:  signals.OperationProcessJob,
		JobID: runnerJob.ID,
	})

	return nil
}
