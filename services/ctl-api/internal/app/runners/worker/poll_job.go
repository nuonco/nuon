package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

const (
	pollJobTimeout time.Duration = time.Minute * 30
	pollJobPeriod  time.Duration = time.Second * 10
)

func (w *Workflows) pollJob(ctx workflow.Context, jobID string) error {
	for {
		job, err := activities.AwaitGetJobByID(ctx, jobID)
		if err != nil {
			return fmt.Errorf("unable to get runner from database: %w", err)
		}

		switch job.Status {
		case app.RunnerJobStatusFailed,
			app.RunnerJobStatusTimedOut,
			app.RunnerJobStatusCancelled,
			app.RunnerJobStatusNotAttempted,
			app.RunnerJobStatusUnknown:
			return fmt.Errorf("failure job status %s", job.Status)
		case app.RunnerJobStatusFinished:
			return nil
		}

		workflow.Sleep(ctx, pollJobPeriod)
	}
}
