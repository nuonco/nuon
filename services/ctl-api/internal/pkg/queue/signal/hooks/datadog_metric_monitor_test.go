package hooks

import (
	"testing"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// TestMonitorMatchesEvent_FailureOnInstallTarget pins the install +
// failure preset path the dashboard's one-click install monitor takes.
// The metric-mode predicate must agree with the event-mode DD query
// (`source:nuon nuon_install_id:<id> nuon_status:failed`) for users
// running both modes side-by-side on the same install — otherwise the
// two monitor modes would fire on different subsets of events.
func TestMonitorMatchesEvent_FailureOnInstallTarget(t *testing.T) {
	m := &app.DatadogManagedMonitor{
		TargetType: app.DatadogManagedMonitorTargetTypeInstall,
		TargetID:   "inst1",
		Preset:     app.DatadogManagedMonitorPresetFailure,
	}
	matching := lifecycleEventData{
		Outcome: &lifecycleOutcome{Status: statusFailed},
	}
	if !monitorMatchesEvent(m, signal.SignalPhaseEvent{}, matching, labels.EventTargets{InstallID: "inst1"}) {
		t.Errorf("expected match for install+failure preset, got false")
	}
	// Wrong install — must not fire even though the outcome is failed.
	if monitorMatchesEvent(m, signal.SignalPhaseEvent{}, matching, labels.EventTargets{InstallID: "inst2"}) {
		t.Errorf("expected no-match for wrong install, got true")
	}
	// Right install, succeeded outcome — must not fire.
	succeeded := lifecycleEventData{Outcome: &lifecycleOutcome{Status: statusSucceeded}}
	if monitorMatchesEvent(m, signal.SignalPhaseEvent{}, succeeded, labels.EventTargets{InstallID: "inst1"}) {
		t.Errorf("expected no-match for succeeded outcome on failure preset, got true")
	}
}

// TestMonitorMatchesEvent_ActionInstallScope mirrors buildMonitorQuery's
// action-target AND install-scope behavior. An action monitor scoped to
// install A must not fire on install B's invocations of the same action.
func TestMonitorMatchesEvent_ActionInstallScope(t *testing.T) {
	m := &app.DatadogManagedMonitor{
		TargetType: app.DatadogManagedMonitorTargetTypeAction,
		TargetID:   "act1",
		InstallID:  "inst1",
		Preset:     app.DatadogManagedMonitorPresetFailure,
	}
	failed := lifecycleEventData{Outcome: &lifecycleOutcome{Status: statusFailed}}

	// Matching: right action + right install + failed.
	if !monitorMatchesEvent(m, signal.SignalPhaseEvent{},
		failed,
		labels.EventTargets{ActionID: "act1", InstallID: "inst1"}) {
		t.Errorf("expected match for action+install scoped failure, got false")
	}
	// Right action, wrong install — must not fire.
	if monitorMatchesEvent(m, signal.SignalPhaseEvent{},
		failed,
		labels.EventTargets{ActionID: "act1", InstallID: "inst2"}) {
		t.Errorf("expected no-match for wrong install on action monitor, got true")
	}
}

// TestMonitorMatchesEvent_DriftPreset confirms drift preset keys off the
// raw drift-detected signal type. The matcher deliberately bypasses
// `data.Kind` because the renderer doesn't emit nuon_kind:drift in v1
// (data.Kind stays "workflow") — keying off SignalType keeps the
// metric path correct regardless of that renderer gap.
func TestMonitorMatchesEvent_DriftPreset(t *testing.T) {
	m := &app.DatadogManagedMonitor{
		TargetType: app.DatadogManagedMonitorTargetTypeInstall,
		TargetID:   "inst1",
		Preset:     app.DatadogManagedMonitorPresetDrift,
	}

	driftEvent := signal.SignalPhaseEvent{SignalType: signalTypeDriftDetected}
	if !monitorMatchesEvent(m, driftEvent, lifecycleEventData{}, labels.EventTargets{InstallID: "inst1"}) {
		t.Errorf("expected drift preset match on drift-detected signal, got false")
	}

	otherEvent := signal.SignalPhaseEvent{SignalType: signalTypeExecuteWorkflow}
	if monitorMatchesEvent(m, otherEvent, lifecycleEventData{}, labels.EventTargets{InstallID: "inst1"}) {
		t.Errorf("expected no-match on non-drift signal for drift preset, got true")
	}
}
