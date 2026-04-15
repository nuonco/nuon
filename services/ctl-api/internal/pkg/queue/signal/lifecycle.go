package signal

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
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
	Operation   string  `json:"operation,omitempty"`
}

type SignalPhaseOutcome struct {
	Status     SignalStatus   `json:"status"`
	ErrMessage string         `json:"err_message,omitempty"`
	Duration   time.Duration  `json:"duration,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type PreExecuteDecision struct {
	Allow    bool           `json:"allow"`
	Reason   string         `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func AllowDecision() PreExecuteDecision {
	return PreExecuteDecision{Allow: true}
}

type SignalLifecycleHook interface {
	Name() string
	Supports(event SignalPhaseEvent) bool
	PreExecute(ctx context.Context, event SignalPhaseEvent) (PreExecuteDecision, error)
	PostExecute(ctx context.Context, event SignalPhaseEvent, outcome SignalPhaseOutcome) error
}

func AsSignalLifecycleHook(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(SignalLifecycleHook)),
		fx.ResultTags(`group:"signal_lifecycle_hooks"`),
	)
}

// Hooks provides standard lifecycle hook support for signals.
// Signals embed this struct to get PreExecute/PostExecute methods.
// Infrastructure fields (QueueSignalID, QueueID, SignalType, OrgID) are
// populated by the handler during initializeState.
// Domain fields (InstallID, ComponentID, Operation) are set by the
// signal in its Init() method.
type Hooks struct {
	QueueSignalID string
	QueueID       string
	SignalType    SignalType
	OrgID         string

	// Domain context — signals set these in Init()
	InstallID   *string
	ComponentID *string
	Operation   string
	LogStreamID string
}

// GetHooks returns a pointer to the embedded Hooks struct.
func (h *Hooks) GetHooks() *Hooks { return h }

// PreExecuteHooks builds a SignalPhaseEvent from the Hooks fields, runs
// pre-execute lifecycle hooks via Temporal activity, and returns
// the decision. Fail-open: if the activity fails, execution is allowed.
func (h *Hooks) PreExecuteHooks(ctx workflow.Context, phase SignalPhase) (*PreExecuteResult, error) {
	event := SignalPhaseEvent{
		QueueSignalID: h.QueueSignalID,
		QueueID:       h.QueueID,
		SignalType:    h.SignalType,
		OrgID:         h.OrgID,
		Phase:         phase,
		InstallID:     h.InstallID,
		ComponentID:   h.ComponentID,
		Operation:     h.Operation,
	}

	resp, err := AwaitRunSignalLifecyclePreExecute(ctx, &RunSignalLifecyclePreExecuteRequest{Event: event})
	if err != nil {
		return &PreExecuteResult{Event: event, Allow: true}, nil
	}

	return &PreExecuteResult{
		Event:  event,
		Allow:  resp.Allow,
		Reason: resp.Reason,
	}, nil
}

// PostExecuteHooks sends post-execute lifecycle hooks as a best-effort operation.
// Uses a disconnected context so hook delivery is not affected by workflow cancellation.
func (h *Hooks) PostExecuteHooks(ctx workflow.Context, event SignalPhaseEvent, outcome SignalPhaseOutcome) {
	dctx, _ := workflow.NewDisconnectedContext(ctx)
	_ = AwaitRunSignalLifecyclePostExecute(dctx, &RunSignalLifecyclePostExecuteRequest{
		Event:   event,
		Outcome: outcome,
	})
}

// BuildOutcome builds a SignalPhaseOutcome from an error and duration.
// If LogStreamID is set, it's included in metadata for cleanup hooks.
func (h *Hooks) BuildOutcome(err error, dur time.Duration) SignalPhaseOutcome {
	outcome := SignalPhaseOutcome{
		Status:   SignalStatusSuccess,
		Duration: dur,
	}
	if h.LogStreamID != "" {
		outcome.Metadata = map[string]any{
			"log_stream_id": h.LogStreamID,
		}
	}
	if err != nil {
		outcome.Status = SignalStatusError
		outcome.ErrMessage = err.Error()
	}
	return outcome
}

// PreExecuteResult is the result of running pre-execute lifecycle hooks.
type PreExecuteResult struct {
	Event  SignalPhaseEvent
	Allow  bool
	Reason string
}
