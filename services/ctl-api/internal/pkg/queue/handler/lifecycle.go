package handler

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func (h *handler) buildSignalPhaseEvent(phase signal.SignalPhase) signal.SignalPhaseEvent {
	event := signal.SignalPhaseEvent{
		QueueSignalID: h.queueSignalID,
		QueueID:       h.queueID,
		Phase:         phase,
	}

	if h.queueSignal != nil {
		event.SignalType = h.queueSignal.Type
		if h.queueSignal.OrgID != nil {
			event.OrgID = *h.queueSignal.OrgID
		}
	}

	if lc, ok := h.sig.(signal.SignalWithLifecycleContext); ok {
		ctx := lc.LifecycleContext()
		if ctx.OrgID != "" {
			event.OrgID = ctx.OrgID
		}
		if ctx.OrgName != "" {
			event.OrgName = ctx.OrgName
		}
		event.InstallID = ctx.InstallID
		if event.InstallID == nil && ctx.OwnerType == "installs" && ctx.OwnerID != "" {
			event.InstallID = &ctx.OwnerID
		}
		event.ComponentID = ctx.ComponentID
		event.SandboxID = ctx.SandboxID
		event.Operation = ctx.Operation
		event.Stage = ctx.Stage
		event.WorkflowID = ctx.WorkflowID
		event.WorkflowType = ctx.WorkflowType
		event.StepID = ctx.StepID
		event.StepName = ctx.StepName
		event.OwnerID = ctx.OwnerID
		event.OwnerType = ctx.OwnerType
		if ctx.OwnerName != "" {
			event.OwnerName = ctx.OwnerName
		}
		if ctx.Metadata != nil {
			event.Metadata = ctx.Metadata
		}
	}

	if h.queueSignal != nil {
		signalCtx := h.queueSignal.SignalContext
		telemetry := signalCtx.WorkflowTelemetry
		if telemetry.OrgID != "" {
			event.OrgID = telemetry.OrgID
		}
		if telemetry.OrgName != "" {
			event.OrgName = telemetry.OrgName
		}
		if telemetry.WorkflowID != "" {
			event.WorkflowID = telemetry.WorkflowID
		}
		if telemetry.WorkflowType != "" {
			event.WorkflowType = telemetry.WorkflowType
		}
		if telemetry.OwnerID != "" {
			event.OwnerID = telemetry.OwnerID
		}
		if telemetry.OwnerType != "" {
			event.OwnerType = telemetry.OwnerType
		}
		if telemetry.OwnerName != "" {
			event.OwnerName = telemetry.OwnerName
		} else if telemetry.InstallName != "" {
			event.OwnerName = telemetry.InstallName
		}
		if event.InstallID == nil && telemetry.InstallID != "" {
			event.InstallID = &telemetry.InstallID
		}
	}

	return event
}

// why: runAfterPhaseSafe runs after-phase hooks as a best-effort operation.
// It uses a disconnected context so that hook delivery is not affected
// by workflow cancellation. Errors are swallowed because after-phase
// hooks must never block or fail the signal execution.
func (h *handler) runAfterPhaseSafe(ctx workflow.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) {
	dctx, _ := workflow.NewDisconnectedContext(ctx)

	_ = signal.AwaitRunSignalLifecycleAfterPhase(dctx, &signal.RunSignalLifecycleAfterPhaseRequest{
		Event:   event,
		Outcome: outcome,
	})
}

func (h *handler) runBeforePhase(ctx workflow.Context, event signal.SignalPhaseEvent) signal.BeforePhaseDecision {
	resp, err := signal.AwaitRunSignalLifecycleBeforePhase(ctx, &signal.RunSignalLifecycleBeforePhaseRequest{
		Event: event,
	})
	if err != nil {
		return signal.AllowPhaseDecision()
	}

	return signal.BeforePhaseDecision{
		Allow:    resp.Allow,
		Reason:   resp.Reason,
		Metadata: resp.Metadata,
	}
}

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
