package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 5m
func (w *Workflows) MngUpdate(ctx workflow.Context, sreq signals.RequestSignal) error {
	runnerJob, err := w.createMngJob(ctx, sreq.ID, app.RunnerJobTypeMngRunnerUpdateVersion, map[string]string{
		"upgrade_type": "mng",
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
