package triggershutdown

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "trigger_shutdown"

type Signal struct {
	signal.Hooks
	RunnerID    string `json:"runner_id"`
	ProcessType string `json:"process_type"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.ProcessType == "" {
		return errors.New("process_type is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get the current active process for this runner and type
	process, err := activities.AwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
		RunnerID:    s.RunnerID,
		ProcessType: s.ProcessType,
	})
	if err != nil {
		// No active process — nothing to shut down
		return nil
	}

	// Noop if the process is not active
	if process.ProcessStatus() != app.RunnerProcessStatusActive {
		return nil
	}

	// Create a shutdown request for this process
	_, err = activities.AwaitCreateRunnerProcessShutdown(ctx, activities.CreateRunnerProcessShutdownRequest{
		RunnerProcessID: process.ID,
		Type:            app.RunnerProcessShutdownTypeGraceful,
		CompositeStatus: app.CompositeStatus{
			Status:                 app.Status(app.RunnerProcessShutdownStatusRequested),
			StatusHumanDescription: "uptime threshold exceeded",
			CreatedAtTS:            workflow.Now(ctx).Unix(),
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to create shutdown for process")
	}

	return nil
}
