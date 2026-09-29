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
		{"unhealthy matches when flag set", signalTypeComponentUnhealthy, componentsCfg(ResourceCfg{ComponentHealth: true}), true},
		{"recovered matches the same flag", signalTypeComponentRecovered, componentsCfg(ResourceCfg{ComponentHealth: true}), true},
		{"unhealthy muted when flag unset", signalTypeComponentUnhealthy, componentsCfg(ResourceCfg{}), false},
		{"recovered muted when flag unset", signalTypeComponentRecovered, componentsCfg(ResourceCfg{}), false},

		{"outcome none does not mute health", signalTypeComponentUnhealthy, componentsCfg(ResourceCfg{ComponentHealth: true, Outcome: OutcomeNone}), true},
		{"unrelated ops filter does not mute health", signalTypeComponentUnhealthy, componentsCfg(ResourceCfg{ComponentHealth: true, Ops: []string{"deploy"}}), true},

		{"drift flag does not deliver health", signalTypeComponentUnhealthy, componentsCfg(ResourceCfg{DriftDetected: true}), false},
		{"health flag does not deliver drift-only subscribers extra events", signalTypeComponentRecovered, componentsCfg(ResourceCfg{DriftDetected: true}), false},

		{"install degraded matches installs flag", signalTypeInstallDegraded, installsCfg(ResourceCfg{InstallDegraded: true}), true},
		{"install degraded muted when flag unset", signalTypeInstallDegraded, installsCfg(ResourceCfg{}), false},
		{"component flag does not deliver the install rollup", signalTypeInstallDegraded, installsCfg(ResourceCfg{ComponentHealth: true}), false},
		{"install flag on components does not deliver component health", signalTypeComponentUnhealthy, componentsCfg(ResourceCfg{InstallDegraded: true}), false},

		{"all events delivers health", signalTypeComponentUnhealthy, AllEvents(), true},
		{"all events delivers the install rollup", signalTypeInstallDegraded, AllEvents(), true},

		{"default opts into component health", signalTypeComponentUnhealthy, Default(), true},
		{"default opts into the install rollup", signalTypeInstallDegraded, Default(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Matches(healthEvent(tt.signalType), nil, nil, tt.in))
		})
	}
}
