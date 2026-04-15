package forceshutdown

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "force_shutdown"

type Signal struct {
	signal.Hooks
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

	// Validate runner exists in database
	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Try process-based shutdown first
	process, err := activities.AwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
		RunnerID:    s.RunnerID,
		ProcessType: string(app.RunnerProcessTypeInstall),
	})
	if err == nil && process != nil && process.ID != "" {
		_, err := activities.AwaitCreateRunnerProcessShutdown(ctx, activities.CreateRunnerProcessShutdownRequest{
			RunnerProcessID: process.ID,
			Type:            app.RunnerProcessShutdownTypeForce,
		})
		if err != nil {
			return errors.Wrap(err, "unable to create process shutdown")
		}
		return nil
	}

	// Fallback: create legacy shutdown job for runners without process tracking
	runnerJob, err := s.createRunnerShutdownJob(ctx, s.RunnerID, map[string]string{
		"shutdown_type": "force",
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

	statusactivities.AwaitUpdateRunnerJobStatusV2(ctx, statusactivities.UpdateRunnerJobStatusV2Request{
		RunnerJobID:       runnerJob.ID,
		Status:            app.RunnerJobStatusAvailable,
		StatusDescription: string(app.RunnerJobStatusAvailable),
	})

	return nil
}

// createRunnerShutdownJob creates a shutdown job for the runner
func (s *Signal) createRunnerShutdownJob(ctx workflow.Context, runnerID string, metadata map[string]string) (*app.RunnerJob, error) {
	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get runner")
	}

	ctx = cctx.SetOrgIDWorkflowContext(ctx, runner.OrgID)
	ctx = cctx.SetAccountIDWorkflowContext(ctx, runner.CreatedByID)

	// Create log stream for shutdown operation
	logStream, err := activities.AwaitCreateLogStreamByOperationID(ctx, runner.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create log stream for shutdown")
	}
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	s.Hooks.LogStreamID = logStream.ID
	// Create shutdown job
	runnerJob, err := activities.AwaitCreateShutdownJob(ctx, &activities.CreateShutdownJobRequest{
		RunnerID:    runner.ID,
		OwnerID:     runner.ID,
		LogStreamID: logStream.ID,
		Metadata:    metadata,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create job")
	}

	return runnerJob, nil
}
