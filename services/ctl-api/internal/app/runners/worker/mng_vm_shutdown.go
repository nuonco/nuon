package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 5m
func (w *Workflows) MngVMShutdown(ctx workflow.Context, sreq signals.RequestSignal) error {
	runnerJob, err := w.createMngJob(ctx, sreq.ID, app.RunnerJobTypeMngVMShutDown, map[string]string{
		"shutdown_type": "vm",
	})
	if err != nil {
		return errors.Wrap(err, "unable to create runner job")
	}

	// automatically set the job status to available, so it is picked up, regardless of what else is in flight.
	if err := activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
		JobID:             runnerJob.ID,
		Status:            app.RunnerJobStatusAvailable,
		StatusDescription: string(app.RunnerJobStatusAvailable),
	}); err != nil {
		return errors.Wrap(err, "unable to update job status")
	}

	return nil
}

func (w *Workflows) createMngJob(ctx workflow.Context, runnerID string, jobType app.RunnerJobType, metadata map[string]string) (*app.RunnerJob, error) {
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

	runnerJob, err := activities.AwaitCreateMngJob(ctx, &activities.CreateMngJobRequest{
		RunnerID:    runner.ID,
		LogStreamID: logStream.ID,
		JobType:     jobType,
		Metadata:    metadata,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create job")
	}

	l.Info("dispatching vm shutdown job to runner mng",
		zap.String("runner_id", runner.ID),
		zap.String("runner_type", string(runner.RunnerGroup.Type)),
	)

	return runnerJob, nil
}
