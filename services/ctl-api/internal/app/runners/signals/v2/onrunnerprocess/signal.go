package onrunnerprocess

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "on_runner_process"

type Signal struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
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
	l.Info("on_runner_process signal received", "runner_id", s.RunnerID, "process_id", s.ProcessID)
	return nil
}
