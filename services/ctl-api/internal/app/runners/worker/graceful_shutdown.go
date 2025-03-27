package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) GracefulShutdown(ctx workflow.Context, sreq signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	status, err := activities.AwaitGetRunnerStatusByID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner status")
	}

	if generics.SliceContains(status, []app.RunnerStatus{
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusProvisioning,
		app.RunnerStatusOffline,
	}) {
		l.Debug("not shutting down runner due to status",
			zap.Any("status", status),
		)
		return nil
	}

	runnerJob, err := w.createRunnerJob(ctx, sreq.ID, map[string]string{
		"shutdown_type": "graceful",
	})
	if err != nil {
		return errors.Wrap(err, "unable to create shutdown runner job")
	}

	// We have to send the signal and then return to allow it to be processed.
	// Waiting for it to complete would deadlock. Not a big deal because
	// we wouldn't do anything differently even if it failed.
	w.evClient.Send(ctx, sreq.ID, &signals.Signal{
		Type:  signals.OperationProcessJob,
		JobID: runnerJob.ID,
	})

	return nil
}

func (w *Workflows) createRunnerJob(ctx workflow.Context, runnerID string, metadata map[string]string) (*app.RunnerJob, error) {
	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get runner")
	}

	ctx = cctx.SetOrgIDWorkflowContext(ctx, runner.OrgID)
	ctx = cctx.SetAccountIDWorkflowContext(ctx, runner.CreatedByID)

	// create a logstream derived from the healthcheck
	logStream, err := activities.AwaitCreateLogStreamByOperationID(ctx, runner.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create log stream for health check")
	}
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get logger")
	}

	runnerJob, err := activities.AwaitCreateShutdownJob(ctx, &activities.CreateShutdownJobRequest{
		RunnerID:    runner.ID,
		OwnerID:     runner.ID,
		LogStreamID: logStream.ID,
		Metadata:    metadata,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create shutdown job")
		return nil, errors.Wrap(err, "unable to create job")
	}

	l.Info("dispatching shutdown job to runner",
		zap.String("runner_id", runner.ID),
		zap.String("runner_type", string(runner.RunnerGroup.Type)),
	)

	return runnerJob, nil
}
