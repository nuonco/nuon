package mngrestart

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "mng-restart"

type Signal struct {
	RunnerID string `json:"runner_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}

	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	runnerJob, err := s.createMngJob(ctx, s.RunnerID, app.RunnerJobTypeMngRunnerRestart, map[string]string{
		"restart_type": "install",
	})
	if err != nil {
		return errors.Wrap(err, "unable to create runner job")
	}

	if err := activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
		JobID:             runnerJob.ID,
		Status:            app.RunnerJobStatusAvailable,
		StatusDescription: string(app.RunnerJobStatusAvailable),
	}); err != nil {
		return errors.Wrap(err, "unable to update job status")
	}

	return nil
}

func (s *Signal) createMngJob(ctx workflow.Context, runnerID string, jobType app.RunnerJobType, metadata map[string]string) (*app.RunnerJob, error) {
	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get runner")
	}

	ctx = cctx.SetOrgIDWorkflowContext(ctx, runner.OrgID)
	ctx = cctx.SetAccountIDWorkflowContext(ctx, runner.CreatedByID)

	logStream, err := activities.AwaitCreateLogStreamByOperationID(ctx, runner.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create log stream")
	}
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	runnerJob, err := activities.AwaitCreateMngJob(ctx, &activities.CreateMngJobRequest{
		RunnerID:    runner.ID,
		LogStreamID: logStream.ID,
		JobType:     jobType,
		Metadata:    metadata,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create job")
	}

	return runnerJob, nil
}
