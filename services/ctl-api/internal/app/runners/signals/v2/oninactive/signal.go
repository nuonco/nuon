package oninactive

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "on_inactive"

type Signal struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`
	Reason    string `json:"reason"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

// LifecycleContext exposes runner identity to the queue dispatcher so that
// lifecycle hooks (webhook, Slack) and the interests classifier can fan this
// event out as `op:runners.inactive` to subscribers. The signal continues to
// do its real work in Execute (terminating the per-process queue); the
// lifecycle context only opts it into the notification pipeline.
//
// OrgID is filled in by the queue handler from the queue signal record, so we
// only need to populate Operation + OwnerID/OwnerType here.
func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		Operation: "runner-inactive",
		OwnerID:   s.RunnerID,
		OwnerType: "runners",
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.ProcessID == "" {
		return errors.New("process_id is required")
	}
	if s.Reason == "" {
		return errors.New("reason is required")
	}

	_, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return errors.Wrap(err, "runner process not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)
	l.Info("on_inactive signal received", "runner_id", s.RunnerID, "process_id", s.ProcessID, "reason", s.Reason)

	// Terminate the process queue: stops all emitters and the queue workflow.
	if err := activities.AwaitTerminateProcessQueue(ctx, activities.TerminateProcessQueueRequest{
		RunnerID:  s.RunnerID,
		ProcessID: s.ProcessID,
	}); err != nil {
		l.Error("failed to terminate process queue", "error", err)
		return errors.Wrap(err, "unable to terminate process queue")
	}

	return nil
}
