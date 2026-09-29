package signal

import (
	"context"
	"time"

	"go.uber.org/fx"
)

type SignalStatus string

const (
	SignalStatusSuccess   SignalStatus = "success"
	SignalStatusError     SignalStatus = "error"
	SignalStatusCancelled SignalStatus = "cancelled"
)

type SignalPhase string

const (
	SignalPhaseValidate SignalPhase = "validate"
	SignalPhaseExecute  SignalPhase = "execute"
	SignalPhaseCancel   SignalPhase = "cancel"
)

type SignalPhaseEvent struct {
	QueueSignalID string      `json:"queue_signal_id"`
	QueueID       string      `json:"queue_id"`
	SignalType    SignalType  `json:"signal_type"`
	OrgID         string      `json:"org_id"`
	OrgName       string      `json:"org_name,omitempty"`
	Phase         SignalPhase `json:"phase"`

	InstallID   *string `json:"install_id,omitempty"`
	ComponentID *string `json:"component_id,omitempty"`
	SandboxID   *string `json:"sandbox_id,omitempty"`
	Operation   string  `json:"operation,omitempty"`
	Stage       string  `json:"stage,omitempty"`

	WorkflowID   string `json:"workflow_id,omitempty"`
	WorkflowType string `json:"workflow_type,omitempty"`

	StepID    string `json:"step_id,omitempty"`
	StepName  string `json:"step_name,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	OwnerType string `json:"owner_type,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`
}

type SignalPhaseOutcome struct {
	Status     SignalStatus   `json:"status"`
	ErrMessage string         `json:"err_message,omitempty"`
	Duration   time.Duration  `json:"duration,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type BeforePhaseDecision struct {
	Allow    bool           `json:"allow"`
	Reason   string         `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func AllowPhaseDecision() BeforePhaseDecision {
	return BeforePhaseDecision{Allow: true}
}

type SignalLifecycleHook interface {
	Name() string
	Supports(event SignalPhaseEvent) bool
	BeforePhase(ctx context.Context, event SignalPhaseEvent) (BeforePhaseDecision, error)
	AfterPhase(ctx context.Context, event SignalPhaseEvent, outcome SignalPhaseOutcome) error
}

type SignalWithLifecycleContext interface {
	Signal

	LifecycleContext() SignalLifecycleContext
}

type SignalLifecycleContext struct {
	OrgID       string  `json:"org_id"`
	OrgName     string  `json:"org_name,omitempty"`
	InstallID   *string `json:"install_id,omitempty"`
	ComponentID *string `json:"component_id,omitempty"`
	SandboxID   *string `json:"sandbox_id,omitempty"`
	Operation   string  `json:"operation"`
	Stage       string  `json:"stage,omitempty"`

	WorkflowID   string `json:"workflow_id,omitempty"`
	WorkflowType string `json:"workflow_type,omitempty"`

	StepID    string `json:"step_id,omitempty"`
	StepName  string `json:"step_name,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	OwnerType string `json:"owner_type,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`
}

type SignalWithMutableLifecycleContext interface {
	SetLifecycleWorkflow(workflowID, workflowType string)
}

type LifecycleBase struct {
	LifecycleWorkflowID   string `json:"lifecycle_workflow_id,omitempty"`
	LifecycleWorkflowType string `json:"lifecycle_workflow_type,omitempty"`
}

func (b *LifecycleBase) SetLifecycleWorkflow(workflowID, workflowType string) {
	b.LifecycleWorkflowID = workflowID
	b.LifecycleWorkflowType = workflowType
}

func AsSignalLifecycleHook(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(SignalLifecycleHook)),
		fx.ResultTags(`group:"signal_lifecycle_hooks"`),
	)
}
