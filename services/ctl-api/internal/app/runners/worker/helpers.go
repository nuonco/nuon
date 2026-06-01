package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) createMngJob(ctx workflow.Context, runnerID string, jobType app.RunnerJobType, metadata map[string]string) (*app.RunnerJob, error) {
	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get runner")
	}

	ctx = cctx.SetOrgIDWorkflowContext(ctx, runner.OrgID)
	ctx = cctx.SetAccountIDWorkflowContext(ctx, runner.CreatedByID)

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
		OwnerID:     runner.RunnerGroup.OwnerID,
		LogStreamID: logStream.ID,
		JobType:     jobType,
		Metadata:    metadata,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create job")
	}

	l.Info(fmt.Sprintf("dispatching %s", jobType),
		zap.String("runner_id", runner.ID),
		zap.String("runner_type", string(runner.RunnerGroup.Type)),
	)

	return runnerJob, nil
}
