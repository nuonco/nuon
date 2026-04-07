package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// @temporal-gen-v2 workflow
// @execution-timeout 10m
// @task-timeout 5m
func (w *Workflows) MngRestart(ctx workflow.Context, sreq signals.RequestSignal) error {
	runnerJob, err := w.createMngJob(ctx, sreq.ID, app.RunnerJobTypeMngRunnerRestart, map[string]string{
		"restart_type": "install",
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

	statusactivities.AwaitUpdateRunnerJobStatusV2(ctx, statusactivities.UpdateRunnerJobStatusV2Request{
		RunnerJobID:       runnerJob.ID,
		Status:            app.RunnerJobStatusAvailable,
		StatusDescription: string(app.RunnerJobStatusAvailable),
	})

	return nil
}
