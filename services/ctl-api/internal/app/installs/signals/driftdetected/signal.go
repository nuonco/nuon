package driftdetected

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "drift-detected"

const installWorkflowStepsOwnerType = "install_workflow_steps"

type Signal struct {
	InstallID         string `json:"install_id"`
	InstallWorkflowID string `json:"install_workflow_id"`
	WorkflowStepID    string `json:"workflow_step_id"`

	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) AutoRetry() bool { return true }
func (s *Signal) MaxRetries() int { return 5 }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	installID := &s.InstallID
	if s.InstallID == "" {
		installID = nil
	}
	return signal.SignalLifecycleContext{
		InstallID:  installID,
		Operation:  "drift-detected",
		WorkflowID: s.InstallWorkflowID,
		StepID:     s.WorkflowStepID,
		OwnerID:    s.InstallID,
		OwnerType:  "installs",
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.WorkflowStepID == "" {
		return errors.New("workflow_step_id is required")
	}
	return nil
}

// why: Execute is intentionally a no-op. The signal exists purely so the queue
// dispatcher emits lifecycle events that the interests classifier can map onto
// the drift_detected slug. Subscribers receive the notification through the
// usual webhook / Slack hook plumbing — there's no business logic to run here.
func (s *Signal) Execute(_ workflow.Context) error {
	return nil
}
