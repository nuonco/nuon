package hooks

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestTelemetrySelectedCarrierEvents(t *testing.T) {
	tests := []struct {
		name       string
		signalType signal.SignalType
		metadata   map[string]any
		event      string
	}{
		{name: "awaiting retry", signalType: signalTypeWorkflowStepAwaitingRetry, metadata: map[string]any{"retry_index": 2}, event: "step.awaiting_retry"},
		{name: "drift", signalType: signalTypeDriftDetected, event: "drift.detected"},
		{name: "config synced", signalType: signalTypeAppConfigSynced, event: "app_config.synced"},
		{name: "config updated", signalType: signalTypeUpdateAppConfig, event: "install.config_updated"},
		{name: "component unhealthy", signalType: signalTypeComponentUnhealthy, metadata: map[string]any{"health": "unhealthy"}, event: "component.unhealthy"},
		{name: "component recovered", signalType: signalTypeComponentRecovered, metadata: map[string]any{"health": "healthy"}, event: "component.recovered"},
		{name: "install degraded", signalType: signalTypeInstallDegraded, metadata: map[string]any{"health": "degraded"}, event: "install.degraded"},
		{name: "install recovered", signalType: signalTypeInstallDegraded, metadata: map[string]any{"health": "healthy"}, event: "install.recovered"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			hook := NewTelemetrySignalLifecycleHook(TelemetryParams{L: zap.New(core)})
			event := signal.SignalPhaseEvent{
				QueueSignalID: "queue-signal-id",
				QueueID:       "queue-id",
				SignalType:    tt.signalType,
				Phase:         signal.SignalPhaseExecute,
				Metadata:      tt.metadata,
			}
			if tt.signalType == signalTypeUpdateAppConfig {
				event.Operation = "update-app-config"
			}
			require.True(t, hook.Supports(event))

			_, err := hook.BeforePhase(context.Background(), event)
			require.NoError(t, err)
			require.NoError(t, hook.AfterPhase(context.Background(), event, signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}))
			require.Len(t, logs.All(), 1)
			require.Equal(t, tt.event, logs.All()[0].ContextMap()["flow_event"])
		})
	}
}

func TestTelemetrySelectedCarrierOmitsFailuresCancellationAndSensitiveMetadata(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	hook := NewTelemetrySignalLifecycleHook(TelemetryParams{L: zap.New(core)})
	event := signal.SignalPhaseEvent{
		SignalType: signalTypeAppConfigSynced,
		Phase:      signal.SignalPhaseExecute,
		Metadata: map[string]any{
			"actor_email": "private@example.com",
			"app_name":    "private-name",
			"message":     "private-message",
		},
	}

	require.NoError(t, hook.AfterPhase(context.Background(), event, signal.SignalPhaseOutcome{Status: signal.SignalStatusError}))
	event.Phase = signal.SignalPhaseCancel
	require.NoError(t, hook.AfterPhase(context.Background(), event, signal.SignalPhaseOutcome{Status: signal.SignalStatusCancelled}))
	require.Empty(t, logs.All())

	event.Phase = signal.SignalPhaseExecute
	require.NoError(t, hook.AfterPhase(context.Background(), event, signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}))
	require.Len(t, logs.All(), 1)
	fields := logs.All()[0].ContextMap()
	require.NotContains(t, fields, "actor_email")
	require.NotContains(t, fields, "app_name")
	require.NotContains(t, fields, "message")
}

func TestTelemetrySelectedCarrierSkipsDryRunAndUnknownHealth(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	hook := NewTelemetrySignalLifecycleHook(TelemetryParams{L: zap.New(core)})
	for _, event := range []signal.SignalPhaseEvent{
		{SignalType: signalTypeUpdateAppConfig, Phase: signal.SignalPhaseExecute},
		{SignalType: signalTypeInstallDegraded, Phase: signal.SignalPhaseExecute},
		{SignalType: signalTypeInstallDegraded, Phase: signal.SignalPhaseExecute, Metadata: map[string]any{"health": "unknown"}},
		{SignalType: signalTypeInstallDegraded, Phase: signal.SignalPhaseExecute, Metadata: map[string]any{"health": "progressing"}},
	} {
		require.NoError(t, hook.AfterPhase(context.Background(), event, signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}))
	}
	require.Empty(t, logs.All())
}

func TestTelemetrySelectedCarrierRepeatedDiagnosticRetainsError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	hook := NewTelemetrySignalLifecycleHook(TelemetryParams{L: zap.New(core)})
	diagnostic := strings.Repeat("failure detail ", 200)
	event := signal.SignalPhaseEvent{
		SignalType: signalTypeWorkflowStepAwaitingRetry, Phase: signal.SignalPhaseExecute,
		QueueSignalID: "queue-signal-id", OrgID: "org-id", StepID: "step-id",
		Metadata: map[string]any{"error": diagnostic, "retry_index": float64(2), "max_retries": 3},
	}
	for range 2 {
		require.NoError(t, hook.AfterPhase(context.Background(), event, signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess, Duration: time.Second}))
	}
	require.Len(t, logs.All(), 2)
	for _, entry := range logs.All() {
		fields := entry.ContextMap()
		require.Equal(t, diagnostic, fields["error"])
		require.Equal(t, "org-id", fields["org_id"])
		require.Equal(t, "queue-signal-id", fields["queue_signal_id"])
		require.EqualValues(t, 2, fields["retry_index"])
		require.EqualValues(t, 3, fields["max_retries"])
		require.EqualValues(t, 1000, fields["signal_duration_ms"])
		require.NotContains(t, fields, "duration_ms")
	}
}
