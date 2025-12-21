package gracefulshutdown

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "graceful_shutdown"

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

	// Validate runner exists in database
	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Check runner status
	status, err := activities.AwaitGetRunnerStatusByID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner status")
	}

	// Skip shutdown for runners in these states
	if generics.SliceContains(status, []app.RunnerStatus{
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusProvisioning,
		app.RunnerStatusOffline,
	}) {
		return nil
	}

	// Create shutdown job
	runnerJob, err := s.createRunnerShutdownJob(ctx, s.RunnerID, map[string]string{
		"shutdown_type": "graceful",
	})
	if err != nil {
		return errors.Wrap(err, "unable to create shutdown runner job")
	}

	// TODO: Implement cross-namespace signal sending for process_job
	// The original code sends a process_job signal via evClient.Send:
	//   w.evClient.Send(ctx, sreq.ID, &signals.Signal{
	//       Type:  signals.OperationProcessJob,
	//       JobID: runnerJob.ID,
	//   })
	// This needs to be implemented using the queue system's cross-namespace
	// signal sending capability (EnqueueSignalToOwner activity)
	_ = runnerJob // Silence unused variable warning until cross-namespace signaling is implemented

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
