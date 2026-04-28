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
	Phase         SignalPhase `json:"phase"`

	InstallID   *string `json:"install_id,omitempty"`
	ComponentID *string `json:"component_id,omitempty"`
	SandboxID   *string `json:"sandbox_id,omitempty"`
	Operation   string  `json:"operation,omitempty"`
	Stage       string  `json:"stage,omitempty"`
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
	InstallID   *string `json:"install_id,omitempty"`
	ComponentID *string `json:"component_id,omitempty"`
	SandboxID   *string `json:"sandbox_id,omitempty"`
	Operation   string  `json:"operation"`
	Stage       string  `json:"stage,omitempty"`
}

func AsSignalLifecycleHook(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(SignalLifecycleHook)),
		fx.ResultTags(`group:"signal_lifecycle_hooks"`),
	)
}
