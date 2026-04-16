package settingschanged

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "settings_changed"

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
	processes, err := activities.AwaitGetActiveRunnerProcesses(ctx, activities.GetActiveRunnerProcessesRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get active runner processes")
	}

	for _, process := range processes {
		if err := activities.AwaitUpdateProcessQueueAndEmittersRequest(ctx, activities.UpdateProcessQueueAndEmittersRequest{
			RunnerID:  s.RunnerID,
			ProcessID: process.ID,
		}); err != nil {
			return errors.Wrapf(err, "unable to update process queue for process %s", process.ID)
		}
	}

	return nil
}
