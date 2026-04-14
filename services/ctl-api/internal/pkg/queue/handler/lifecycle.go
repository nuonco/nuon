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

// logStreamMetadata returns a metadata map containing the log stream ID
// if the signal implements SignalWithLogStream and has a non-empty ID.
// Returns nil otherwise.
func logStreamMetadata(sig signal.Signal) map[string]any {
	if ls, ok := sig.(signal.SignalWithLogStream); ok {
		if id := ls.LogStreamID(); id != "" {
			return map[string]any{
				"log_stream_id": id,
			}
		}
	}
	return nil
}

// buildOutcome builds a SignalPhaseOutcome from an error and duration.
// If the signal implements SignalWithLogStream, the log stream ID is
// included in metadata so that lifecycle hooks can close it.
func buildOutcome(sig signal.Signal, err error, dur time.Duration) signal.SignalPhaseOutcome {
	outcome := signal.SignalPhaseOutcome{
		Status:   signal.SignalStatusSuccess,
		Duration: dur,
		Metadata: logStreamMetadata(sig),
	}
	if err != nil {
		outcome.Status = signal.SignalStatusError
		outcome.ErrMessage = err.Error()
	}

	return outcome
}
