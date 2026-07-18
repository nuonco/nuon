package workflowstepawaitingretry

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// SignalType identifies a workflow step parked awaiting manual retry.
const SignalType signal.SignalType = "workflow-step-awaiting-retry"

// installSignalsQueueName mirrors the constant in
// services/ctl-api/internal/app/installs/helpers. Duplicated here as a
// literal to avoid an import cycle (helpers imports signals via fx wiring).
const installSignalsQueueName = "install-signals"

type Signal struct {
	OrgID        string `json:"org_id"`
	InstallID    string `json:"install_id"`
	WorkflowID   string `json:"workflow_id"`
	WorkflowType string `json:"workflow_type"`
	StepID       string `json:"step_id"`
	StepName     string `json:"step_name"`

	ErrMessage string `json:"err_message,omitempty"`

	RetryIndex int `json:"retry_index"`
	MaxRetries int `json:"max_retries,omitempty"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	installID := &s.InstallID
	if s.InstallID == "" {
		installID = nil
	}
	return signal.SignalLifecycleContext{
		OrgID:        s.OrgID,
		InstallID:    installID,
		Operation:    "workflow-step-awaiting-retry",
		WorkflowID:   s.WorkflowID,
		WorkflowType: s.WorkflowType,
		StepID:       s.StepID,
		OwnerID:      s.InstallID,
		OwnerType:    "installs",
		Metadata: map[string]any{
			"awaiting_retry":         true,
			"manual_action_required": true,
			"terminal":               false,
			"error":                  s.ErrMessage,
			"step_name":              s.StepName,
			"retry_index":            s.RetryIndex,
			"max_retries":            s.MaxRetries,
		},
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	if s.StepID == "" {
		return errors.New("step_id is required")
	}
	return nil
}

func (s *Signal) Execute(_ workflow.Context) error {
	return nil
}
