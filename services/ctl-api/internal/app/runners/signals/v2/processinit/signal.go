package processinit

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "process_init"

type Signal struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SleepAfter = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SleepAfter() time.Duration {
	return 0
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.ProcessID == "" {
		return errors.New("process_id is required")
	}

	_, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return errors.Wrap(err, "runner process not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Get process and verify it's in pending state
	process, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner process")
	}

	// Only transition pending processes to active
	if process.ProcessStatus() != app.RunnerProcessStatus(app.StatusPending) {
		l.Info("process not pending, skipping", "process_id", s.ProcessID, "status", string(process.ProcessStatus()))
		return nil
	}

	// Transition process from pending to active
	if _, err := activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
		ProcessID:         s.ProcessID,
		Status:            app.RunnerProcessStatusActive,
		StatusDescription: "process initialized",
	}); err != nil {
		return errors.Wrap(err, "unable to update process status")
	}

	// Update runner status to active
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusActive,
		StatusDescription: "process initialized",
	}); err != nil {
		return errors.Wrap(err, "unable to update runner status")
	}
	statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusActive,
		StatusDescription: "process initialized",
	})

	l.Info("process initialized", "runner_id", s.RunnerID, "process_id", s.ProcessID)

	return nil
}
