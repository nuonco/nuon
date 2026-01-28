package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 5m
func (w *Workflows) MngFetchToken(ctx workflow.Context, sreq signals.RequestSignal) error {
	// NOTE(fd): platform will need to be set dynamically when we add support for other clouds
	runnerJob, err := w.createMngJob(ctx, sreq.ID, app.RunnerJobTypeMngFetchToken, map[string]string{
		"platform": "aws",
	})
	if err != nil {
		return errors.Wrap(err, "unable to create runner job")
	}

	// NOTE(fd): this one does not have to become available immediately btw
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
