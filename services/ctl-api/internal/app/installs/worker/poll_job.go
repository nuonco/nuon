package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

const (
	pollJobTimeout time.Duration = time.Minute * 30
	pollJobPeriod  time.Duration = time.Second * 10
)

func (w *Workflows) pollJob(ctx workflow.Context, jobID string) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

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

		if job.Status == app.RunnerJobStatusQueued {
			queued, err := activities.AwaitGetRunnerJobQueueByJobID(ctx, jobID)
			if err != nil {
				return errors.Wrap(err, "unable to get runner job queue")
			}

			fields := make([]zapcore.Field, 0)
			for idx, qj := range queued {
				fields = append(fields, zap.String(fmt.Sprintf("job.%d - %s", idx, qj.Type), qj.ID))
			}
			l.Warn(fmt.Sprintf("waiting on %d jobs queued before being attempted", len(queued)), fields...)
		}

		workflow.Sleep(ctx, pollJobPeriod)
	}
}
