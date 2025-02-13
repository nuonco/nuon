package job

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job/activities"
)

func (j *Workflows) logJobQueue(ctx workflow.Context, jobID string) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "expected a log stream in the context to poll job")
	}

	queued, err := activities.AwaitPkgWorkflowsJobGetRunnerJobQueueByJobID(ctx, jobID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job queue")
	}

	fields := make([]zapcore.Field, 0)
	fields = append(fields, zap.Int("queue-depth", len(queued)))
	for idx, qj := range queued {
		fields = append(fields, zap.String(fmt.Sprintf("job.%d - %s", idx, qj.Type), qj.ID))
	}

	switch len(queued) {
	case 0:
	case activities.LimitQueueSize:
		l.Warn(fmt.Sprintf("waiting on at least %d jobs queued before being attempted", activities.LimitQueueSize), fields...)
	default:
		l.Warn(fmt.Sprintf("waiting on %d jobs queued before being attempted", len(queued)), fields...)
	}

	return nil
}
