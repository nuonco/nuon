package signal

import (
	"context"
	"time"

	"go.uber.org/fx"
)

// SignalStatus mirrors app.Status to avoid an import cycle between signal and app.
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

	// WorkflowID and WorkflowType identify the install workflow that owns this
	// signal, when applicable. Sourced from the signal's lifecycle context
	// (set at construction time via SignalWithMutableLifecycleContext). Empty
	// for signals not owned by a workflow step (e.g. install-created).
	WorkflowID   string `json:"workflow_id,omitempty"`
	WorkflowType string `json:"workflow_type,omitempty"`

	// StepID, OwnerID, OwnerType identify the workflow step (when applicable)
	// and the entity that owns the queue (e.g. install). Populated by signals
	// that implement SignalWithLifecycleContext. Used by lifecycle hooks to
	// reason about workflow/step lifecycle without having to know the inner
	// signal taxonomy.
	StepID    string `json:"step_id,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	OwnerType string `json:"owner_type,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
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

// SignalWithLifecycleContext is an optional interface that signals can implement
// to provide rich context for lifecycle hook events.
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

	// WorkflowID and WorkflowType identify the install workflow that owns this
	// signal, when applicable. Populated at signal-construction time via
	// SignalWithMutableLifecycleContext (see LifecycleBase). Empty for signals
	// not owned by a workflow step.
	WorkflowID   string `json:"workflow_id,omitempty"`
	WorkflowType string `json:"workflow_type,omitempty"`

	// StepID, OwnerID, OwnerType identify the workflow step (when applicable)
	// and the entity owning the queue. Populated for the workflow-level
	// (execute-workflow) and step-level (execute-workflow-step) primitives so
	// hooks can emit workflow/workflow_step lifecycle events without leaking
	// inner signal taxonomy.
	StepID    string `json:"step_id,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	OwnerType string `json:"owner_type,omitempty"`
	// OwnerName is the human-readable owner label resolved by the signal at
	// Validate() time. Stamped onto SignalPhaseEvent by the queue handler so
	// lifecycle hooks (e.g. webhook) can emit owner_name without a per-event
	// DB lookup.
	OwnerName string `json:"owner_name,omitempty"`
}

// SignalWithMutableLifecycleContext is an optional interface signals can
// implement to receive workflow identity at construction time. The simplest way
// to implement it is to embed LifecycleBase.
type SignalWithMutableLifecycleContext interface {
	SetLifecycleWorkflow(workflowID, workflowType string)
}

// LifecycleBase can be embedded in signal structs to satisfy
// SignalWithMutableLifecycleContext. The fields are exported with distinctive
// names (prefixed `Lifecycle*`) to avoid colliding with existing signal fields
// and so they round-trip cleanly through JSON serialization into the
// QueueSignal row.
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
