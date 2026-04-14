package handler

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// PhaseFunc is the inner operation for a signal phase (e.g. sig.Validate, sig.Execute).
type PhaseFunc func(ctx workflow.Context) error

// withLifecycle wraps a PhaseFunc with before/after lifecycle hooks.
//
//   - Builds the event from handler state
//   - Runs before-phase hooks (fail-open); returns *SignalErrBlocked if blocked
//   - Times the inner function
//   - Runs after-phase hooks (best-effort, disconnected context)
//   - Returns the inner function's error unchanged
func (h *handler) withLifecycle(ctx workflow.Context, phase signal.SignalPhase, fn PhaseFunc) error {
	event := h.buildSignalPhaseEvent(phase)

	decision := h.runBeforePhase(ctx, event)
	if !decision.Allow {
		return &signal.SignalErrBlocked{Phase: phase, Reason: decision.Reason}
	}

	start := workflow.Now(ctx)
	err := fn(ctx)
	dur := workflow.Now(ctx).Sub(start)

	h.runAfterPhaseSafe(ctx, event, buildOutcome(h.sig, err, dur))
	return err
}

// afterLifecycle fires after-phase hooks only, for phases that have no inner
// function to wrap (e.g. cancel, which IS the operation rather than wrapping one).
func (h *handler) afterLifecycle(ctx workflow.Context, phase signal.SignalPhase, outcome signal.SignalPhaseOutcome) {
	event := h.buildSignalPhaseEvent(phase)
	h.runAfterPhaseSafe(ctx, event, outcome)
}
