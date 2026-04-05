package processuptimecheck

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "process_uptime_check"

// Default uptime thresholds before triggering a shutdown
const (
	DefaultMngUptimeThreshold     = 24 * time.Hour
	DefaultInstallUptimeThreshold = 12 * time.Hour
)

type Signal struct {
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
		// No active process — nothing to check
		return nil
	}

	// Noop if the process is not active (offline, inactive, shut-down, etc.)
	if process.ProcessStatus() != app.RunnerProcessStatusActive {
		return nil
	}

	if process.StartedAt == nil {
		return nil
	}

	// Determine uptime threshold based on process type
	threshold := DefaultInstallUptimeThreshold
	if process.Type == app.RunnerProcessTypeMng {
		threshold = DefaultMngUptimeThreshold
	}

	// Check if process has exceeded uptime threshold
	uptime := workflow.Now(ctx).Sub(*process.StartedAt)
	if uptime < threshold {
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
		return errors.Wrap(err, "unable to create shutdown for process uptime check")
	}

	return nil
}
