package interests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func healthEvent(signalType signal.SignalType) signal.SignalPhaseEvent {
	return signal.SignalPhaseEvent{
		SignalType: signalType,
		Phase:      signal.SignalPhaseExecute,
	}
}

// Health carriers are emitted outside any workflow, so they must classify with
// no WorkflowType, no StepID, and no DB — the paths every other signal relies on.
//
// Component health (unhealthy/recovered) slugs are still produced by Classify
// for diagnostic/webhook-payload purposes, but the matcher suppresses them
// unconditionally (see match.go). Install degraded slugs are likewise still
// classified but suppressed by the matcher.
func TestClassifyComponentHealthSlugs(t *testing.T) {
	tests := []struct {
		name       string
		signalType signal.SignalType
		want       []string
	}{
		{
			name:       "component unhealthy",
			signalType: signalTypeComponentUnhealthy,
			want:       []string{"resource:components", "op:components.health", "event:component.unhealthy"},
		},
		{
			name:       "component recovered",
			signalType: signalTypeComponentRecovered,
			want:       []string{"resource:components", "op:components.health", "event:component.recovered"},
		},
		{
			name:       "install degraded",
			signalType: signalTypeInstallDegraded,
			want:       []string{"resource:installs", "op:installs.health", "event:install.degraded"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Classify(healthEvent(tt.signalType), nil, nil))
		})
	}
}

// Component health and install degraded notifications are retired — the
// matcher suppresses them unconditionally for all subscribers, including
// AllEvents and existing subscriptions that still carry the flags.
func TestMatchesComponentHealth(t *testing.T) {
	componentsCfg := func(cfg ResourceCfg) Interests {
		return Interests{Resources: map[ResourceKind]ResourceCfg{ResourceComponents: cfg}}
	}
	installsCfg := func(cfg ResourceCfg) Interests {
		return Interests{Resources: map[ResourceKind]ResourceCfg{ResourceInstalls: cfg}}
	}

	tests := []struct {
		name       string
		signalType signal.SignalType
		in         Interests
		want       bool
	}{
		// Component health is suppressed unconditionally — no config matches.
		{"unhealthy suppressed even with flag set", signalTypeComponentUnhealthy, componentsCfg(ResourceCfg{ComponentHealth: true}), false},
		{"recovered suppressed even with flag set", signalTypeComponentRecovered, componentsCfg(ResourceCfg{ComponentHealth: true}), false},
		{"unhealthy suppressed under all events", signalTypeComponentUnhealthy, AllEvents(), false},
		{"unhealthy suppressed under default", signalTypeComponentUnhealthy, Default(), false},

		// Install degraded is suppressed unconditionally — no config matches.
		{"install degraded suppressed even with flag set", signalTypeInstallDegraded, installsCfg(ResourceCfg{InstallDegraded: true}), false},
		{"install degraded suppressed when flag unset", signalTypeInstallDegraded, installsCfg(ResourceCfg{}), false},
		{"install degraded suppressed under all events", signalTypeInstallDegraded, AllEvents(), false},
		{"install degraded suppressed under default", signalTypeInstallDegraded, Default(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Matches(healthEvent(tt.signalType), nil, nil, tt.in))
		})
	}
}
