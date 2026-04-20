package handler

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// buildSignalPhaseEvent creates a SignalPhaseEvent from the handler's current state.
// If the signal implements SignalWithLifecycleContext, it enriches the event
// with install/component/operation metadata.
func (h *handler) buildSignalPhaseEvent(phase signal.SignalPhase) signal.SignalPhaseEvent {
	event := signal.SignalPhaseEvent{
		QueueSignalID: h.queueSignalID,
		QueueID:       h.queueID,
		Phase:         phase,
	}

	// populate from loaded queue signal state
	if h.queueSignal != nil {
		event.SignalType = h.queueSignal.Type
		if h.queueSignal.OrgID != nil {
			event.OrgID = *h.queueSignal.OrgID
		}
	}

	// enrich from signal if it implements the optional lifecycle context interface
	if lc, ok := h.sig.(signal.SignalWithLifecycleContext); ok {
		ctx := lc.LifecycleContext()
		if ctx.OrgID != "" {
			event.OrgID = ctx.OrgID
		}
		event.InstallID = ctx.InstallID
		event.ComponentID = ctx.ComponentID
		event.Operation = ctx.Operation
	}

	return event
}

// runAfterPhaseSafe runs after-phase hooks as a best-effort operation.
// It uses a disconnected context so that hook delivery is not affected
// by workflow cancellation. Errors are swallowed because after-phase
// hooks must never block or fail the signal execution.
func (h *handler) runAfterPhaseSafe(ctx workflow.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) {
	// Skip the activity round-trip if no registered hook would handle this event.
	// This avoids ~1s of Temporal orchestration overhead per signal phase.
	if !hookPhaseMayRun(event.Phase) {
		return
	}

	// use a disconnected context so cancellation doesn't prevent hook delivery
	dctx, _ := workflow.NewDisconnectedContext(ctx)

	_ = signal.AwaitRunSignalLifecycleAfterPhase(dctx, &signal.RunSignalLifecycleAfterPhaseRequest{
		Event:   event,
		Outcome: outcome,
	})
}

// runBeforePhase runs before-phase hooks and returns the decision.
// If hook execution fails, it returns an allow decision (fail-open).
func (h *handler) runBeforePhase(ctx workflow.Context, event signal.SignalPhaseEvent) signal.BeforePhaseDecision {
	// Skip the activity round-trip if no registered hook would handle this event.
	// This avoids ~1s of Temporal orchestration overhead per signal phase.
	if !hookPhaseMayRun(event.Phase) {
		return signal.AllowPhaseDecision()
	}

	resp, err := signal.AwaitRunSignalLifecycleBeforePhase(ctx, &signal.RunSignalLifecycleBeforePhaseRequest{
		Event: event,
	})
	if err != nil {
		// fail-open: if hooks fail to run, allow execution to continue
		return signal.AllowPhaseDecision()
	}

	return signal.BeforePhaseDecision{
		Allow:    resp.Allow,
		Reason:   resp.Reason,
		Metadata: resp.Metadata,
	}
}

// hookPhaseMayRun returns true when at least one registered lifecycle hook
// could plausibly fire for the given phase. Today the only registered hook
// is WebhookSignalLifecycleHook, which explicitly skips the validate phase
// in its Supports() method, so calling the activity for validate only incurs
// Temporal overhead with no work done.
//
// If a future hook needs to run in validate phase, update this check.
func hookPhaseMayRun(phase signal.SignalPhase) bool {
	return phase != signal.SignalPhaseValidate
}

// outcomeFromError builds a SignalPhaseOutcome from an error and duration.
func outcomeFromError(err error, dur time.Duration) signal.SignalPhaseOutcome {
	if err != nil {
		return signal.SignalPhaseOutcome{
			Status:     signal.SignalStatusError,
			ErrMessage: err.Error(),
			Duration:   dur,
		}
	}

	return signal.SignalPhaseOutcome{
		Status:   signal.SignalStatusSuccess,
		Duration: dur,
	}
}
