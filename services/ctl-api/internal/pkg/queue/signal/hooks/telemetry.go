package hooks

import (
	"context"
	"strconv"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
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
	return telemetryKind(event.SignalType) != telemetryKindSignal || selectedDiagnosticEvent(event.SignalType) != ""
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
	if flowEvent := selectedDiagnosticEvent(event.SignalType); flowEvent != "" {
		if event.Phase != signal.SignalPhaseExecute {
			return nil
		}
		if event.SignalType == signalTypeUpdateAppConfig {
			if event.Operation != "update-app-config" {
				return nil
			}
			if outcome.Status == signal.SignalStatusError {
				h.emit(ctx, "install.config_update_failed", event, &outcome)
				return nil
			}
		}
		if outcome.Status != signal.SignalStatusSuccess {
			return nil
		}
		if event.SignalType == signalTypeInstallDegraded {
			switch boundedHealth(event.Metadata["health"]) {
			case "healthy":
				flowEvent = "install.recovered"
			case "unhealthy", "degraded":
			default:
				return nil
			}
		}
		h.emit(ctx, flowEvent, event, &outcome)
		return nil
	}

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
		return nil
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

func selectedDiagnosticEvent(signalType signal.SignalType) string {
	switch signalType {
	case signalTypeWorkflowStepAwaitingRetry:
		return "step.awaiting_retry"
	case signalTypeDriftDetected:
		return "drift.detected"
	case signalTypeAppConfigSynced:
		return "app_config.synced"
	case signalTypeUpdateAppConfig:
		return "install.config_updated"
	case signalTypeComponentUnhealthy:
		return "component.unhealthy"
	case signalTypeComponentRecovered:
		return "component.recovered"
	case signalTypeInstallDegraded:
		return "install.degraded"
	default:
		return ""
	}
}

func boundedHealth(value any) string {
	health, _ := value.(string)
	switch health {
	case "healthy", "degraded", "unhealthy", "unknown", "progressing", "not-applicable":
		return health
	default:
		return ""
	}
}

func boundedInteger(value any) (int64, bool) {
	switch value := value.(type) {
	case int:
		return int64(value), true
	case int64:
		return value, true
	case float64:
		return int64(value), value == float64(int64(value))
	case string:
		parsed, err := strconv.ParseInt(value, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func (h *TelemetrySignalLifecycleHook) emit(ctx context.Context, flowEvent string, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) {
	fields := []zap.Field{
		zap.String("flow_event", flowEvent),
		zap.String("source", "signal_lifecycle"),
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("queue_id", event.QueueID),
		zap.String("org_id", event.OrgID),
		zap.String("org_name", event.OrgName),
		zap.String("install_id", eventInstallID(event)),
		zap.String("workflow_id", event.WorkflowID),
		zap.String("workflow_type", event.WorkflowType),
		zap.String("owner_id", event.OwnerID),
		zap.String("owner_type", event.OwnerType),
		zap.String("operation", event.Operation),
		zap.String("stage", event.Stage),
		zap.String("signal_type", string(event.SignalType)),
		zap.String("phase", string(event.Phase)),
	}
	if event.OwnerType == plugins.TableNameOf[app.Install]() {
		fields = append(fields, zap.String("install_name", event.OwnerName))
	}
	if event.ComponentID != nil {
		fields = append(fields, zap.String("component_id", *event.ComponentID))
	}
	if event.SandboxID != nil {
		fields = append(fields, zap.String("sandbox_id", *event.SandboxID))
	}
	if event.StepID != "" {
		fields = append(fields, zap.String("step_id", event.StepID))
	}
	if event.StepName != "" {
		fields = append(fields, zap.String("step_name", event.StepName))
	}
	for _, key := range []string{"step_group_id", "step_idx", "group_idx", "group_retry_idx"} {
		if value, ok := event.Metadata[key]; ok {
			fields = append(fields, zap.Any(key, value))
		}
	}
	for _, key := range []string{"health", "previous_health", "install_health", "install_previous_health"} {
		if value := boundedHealth(event.Metadata[key]); value != "" {
			fields = append(fields, zap.String(key, value))
		}
	}
	for _, key := range []string{"retry_index", "max_retries", "unhealthy_component_count", "degraded_component_count"} {
		if value, ok := boundedInteger(event.Metadata[key]); ok {
			fields = append(fields, zap.Int64(key, value))
		}
	}

	failed := false
	if outcome != nil {
		if selectedDiagnosticEvent(event.SignalType) != "" {
			fields = append(fields, zap.Int64("signal_duration_ms", outcome.Duration.Milliseconds()))
		} else if outcome.Duration > 0 {
			fields = append(fields, zap.Int64("duration_ms", outcome.Duration.Milliseconds()))
		}
		if outcome.Status == signal.SignalStatusError {
			failed = true
			if outcome.ErrMessage != "" {
				fields = append(fields, zap.String("error", outcome.ErrMessage))
			}
		} else if event.SignalType == signalTypeWorkflowStepAwaitingRetry {
			if value, ok := event.Metadata["error"].(string); ok && value != "" {
				fields = append(fields, zap.String("error", value))
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
		return kind + ".failed"
	}
	return kind + ".completed"
}

func eventInstallID(event signal.SignalPhaseEvent) string {
	if event.InstallID != nil {
		return *event.InstallID
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
