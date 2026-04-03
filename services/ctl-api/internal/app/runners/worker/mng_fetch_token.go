package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen-v2 workflow
// @execution-timeout 10m
// @task-timeout 5m
func (w *Workflows) MngFetchToken(ctx workflow.Context, sreq signals.RequestSignal) error {
	runner, err := activities.AwaitGetByRunnerID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner for platform detection")
	}

	platform := string(runner.RunnerGroup.Platform)
	if platform == "" {
		platform = "aws"
	}

	runnerJob, err := w.createMngJob(ctx, sreq.ID, app.RunnerJobTypeMngFetchToken, map[string]string{
		"platform": platform,
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
