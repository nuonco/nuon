package hooks

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// duplicated as a string constant to avoid importing the installs/signals tree
// (which would introduce an import cycle), mirroring how webhook.go declares its
// signal types.
const (
	signalTypeGenerateWorkflowSteps    signal.SignalType = "generate-workflow-steps"
	signalTypeExecuteWorkflowStepGroup signal.SignalType = "execute-workflow-step-group"
)

const (
	telemetryKindWorkflow       = "workflow"
	telemetryKindStep           = "step"
	telemetryKindStepGroup      = "step_group"
	telemetryKindApproval       = "approval"
	telemetryKindStepGeneration = "step_generation"
	telemetryKindSignal         = "signal"
)

const telemetryMsg = "flow telemetry"

type TelemetryParams struct {
	fx.In

	L *zap.Logger `optional:"true"`
}

// TelemetrySignalLifecycleHook emits one structured log line per flow lifecycle
// transition, reading only fields the originating signal already propagated (no
// database access).
type TelemetrySignalLifecycleHook struct {
	l *zap.Logger
}

var _ signal.SignalLifecycleHook = (*TelemetrySignalLifecycleHook)(nil)

func NewTelemetrySignalLifecycleHook(params TelemetryParams) *TelemetrySignalLifecycleHook {
	logger := params.L
	if logger == nil {
		logger = zap.NewNop()
	}

	return &TelemetrySignalLifecycleHook{l: logger}
}

func (h *TelemetrySignalLifecycleHook) Name() string {
	return "flow_lifecycle_telemetry"
}

func (h *TelemetrySignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	return telemetryKind(event.SignalType) != telemetryKindSignal
}

func (h *TelemetrySignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	if event.Phase != signal.SignalPhaseExecute {
		return signal.AllowPhaseDecision(), nil
	}

	switch telemetryKind(event.SignalType) {
	case telemetryKindWorkflow:
		h.emit(ctx, "workflow.started", event, nil)
	case telemetryKindStepGeneration:
		h.emit(ctx, "step_generation.started", event, nil)
	case telemetryKindStepGroup:
		h.emit(ctx, "step_group.started", event, nil)
	case telemetryKindStep:
		h.emit(ctx, "step.started", event, nil)
	}

	return signal.AllowPhaseDecision(), nil
}

func (h *TelemetrySignalLifecycleHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	if event.Phase == signal.SignalPhaseValidate {
		return nil
	}

	kind := telemetryKind(event.SignalType)
	errored := outcome.Status == signal.SignalStatusError

	if event.Phase == signal.SignalPhaseCancel {
		switch kind {
		case telemetryKindWorkflow, telemetryKindStep, telemetryKindStepGroup, telemetryKindStepGeneration:
			h.emit(ctx, kind+".cancelled", event, &outcome)
		}
		return nil
	}

	switch kind {
	case telemetryKindWorkflow:
		// Terminal error comes from the status-update activity: a workflow can
		// fail while its execute signal is still parked, so the outcome here is
		// not a reliable error source.
		if !errored {
			h.emit(ctx, "workflow.completed", event, &outcome)
		}
	case telemetryKindStepGeneration:
		h.emit(ctx, terminalEvent("step_generation", errored), event, &outcome)
	case telemetryKindStepGroup:
		h.emit(ctx, terminalEvent("step_group", errored), event, &outcome)
	case telemetryKindStep:
		// Terminal error comes from the status-update activity: a failed step
		// can park (awaiting retry) without its signal returning.
		if !errored {
			h.emit(ctx, "step.completed", event, &outcome)
		}
	case telemetryKindApproval:
		if errored {
			return nil
		}
		switch event.SignalType {
		case signalTypeWorkflowStepApprovalRequest:
			h.emit(ctx, "step.awaiting_approval", event, nil)
		case signalTypeWorkflowStepApprovalResponse:
			h.emit(ctx, "step.approval_resolved", event, nil)
		}
	}

	return nil
}

func (h *TelemetrySignalLifecycleHook) emit(ctx context.Context, flowEvent string, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) {
	fields := []zap.Field{
		zap.String("flow_event", flowEvent),
		zap.String("install_id", eventInstallID(event)),
		zap.String("workflow_id", event.WorkflowID),
		zap.String("workflow_type", event.WorkflowType),
		zap.String("operation", event.Operation),
		zap.String("signal_type", string(event.SignalType)),
		zap.String("phase", string(event.Phase)),
	}
	if event.StepID != "" {
		fields = append(fields, zap.String("step_id", event.StepID))
	}
	if event.StepName != "" {
		fields = append(fields, zap.String("step_name", event.StepName))
	}
	if v, ok := event.Metadata["step_group_id"]; ok {
		fields = append(fields, zap.Any("step_group_id", v), zap.Any("group_idx", event.Metadata["group_idx"]))
	}

	failed := false
	if outcome != nil {
		if outcome.Duration > 0 {
			fields = append(fields, zap.Int64("duration_ms", outcome.Duration.Milliseconds()))
		}
		if outcome.Status == signal.SignalStatusError {
			failed = true
			if outcome.ErrMessage != "" {
				fields = append(fields, zap.String("error", outcome.ErrMessage))
			}
		}
	}

	l := cctx.GetLogger(ctx, h.l)
	if failed {
		l.Error(telemetryMsg, fields...)
		return
	}
	l.Info(telemetryMsg, fields...)
}

func terminalEvent(kind string, errored bool) string {
	if errored {
		if kind == "step" {
			return "step.errored"
		}
		return kind + ".failed"
	}
	return kind + ".completed"
}

func eventInstallID(event signal.SignalPhaseEvent) string {
	if event.InstallID != nil && *event.InstallID != "" {
		return *event.InstallID
	}
	if event.OwnerType == "installs" {
		return event.OwnerID
	}
	return ""
}

func telemetryKind(st signal.SignalType) string {
	switch st {
	case signalTypeExecuteWorkflow:
		return telemetryKindWorkflow
	case signalTypeExecuteWorkflowStep:
		return telemetryKindStep
	case signalTypeExecuteWorkflowStepGroup:
		return telemetryKindStepGroup
	case signalTypeWorkflowStepApprovalRequest, signalTypeWorkflowStepApprovalResponse:
		return telemetryKindApproval
	case signalTypeGenerateWorkflowSteps:
		return telemetryKindStepGeneration
	default:
		return telemetryKindSignal
	}
}
