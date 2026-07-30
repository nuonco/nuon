// Package componenthealthnotify holds notification-only carrier signals for the
// component health axis. Execute is a no-op — they exist only so the dispatcher
// emits events for subscribers, and carry no WorkflowID/StepID.
package componenthealthnotify

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	ComponentUnhealthySignalType signal.SignalType = "component-unhealthy"
	ComponentRecoveredSignalType signal.SignalType = "component-recovered"
	InstallDegradedSignalType    signal.SignalType = "install-degraded"
)

// Metadata keys carried through to the webhook payload and Slack renderers.
const (
	MetadataKeyHealth         = "health"
	MetadataKeyPreviousHealth = "previous_health"
	MetadataKeyMessage        = "message"
	MetadataKeyComponentName  = "component_name"
	MetadataKeyResourceKind   = "root_resource_kind"
	MetadataKeyResourceNS     = "root_resource_namespace"
	MetadataKeyResourceName   = "root_resource_name"
	MetadataKeyUnhealthyCount = "unhealthy_component_count"
	MetadataKeyDegradedCount  = "degraded_component_count"
)

// ComponentSignal is the shared body of the two component-level health
// notifications; Health/PreviousHealth let a renderer say "went from X to Y".
type ComponentSignal struct {
	InstallID          string `json:"install_id"`
	InstallName        string `json:"install_name"`
	InstallComponentID string `json:"install_component_id"`
	ComponentID        string `json:"component_id"`
	ComponentName      string `json:"component_name"`

	Health         string `json:"health"`
	PreviousHealth string `json:"previous_health"`
	Message        string `json:"message"`

	RootResourceKind      string `json:"root_resource_kind"`
	RootResourceNamespace string `json:"root_resource_namespace"`
	RootResourceName      string `json:"root_resource_name"`
}

func (s *ComponentSignal) validate() error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.InstallComponentID == "" {
		return errors.New("install_component_id is required")
	}
	if s.Health == "" {
		return errors.New("health is required")
	}
	return nil
}

func (s *ComponentSignal) lifecycleContext(operation string) signal.SignalLifecycleContext {
	installID := &s.InstallID
	componentID := &s.ComponentID
	if s.ComponentID == "" {
		componentID = nil
	}

	return signal.SignalLifecycleContext{
		InstallID:   installID,
		ComponentID: componentID,
		Operation:   operation,
		OwnerID:     s.InstallID,
		OwnerType:   "installs",
		OwnerName:   s.InstallName,
		Metadata: map[string]any{
			MetadataKeyHealth:         s.Health,
			MetadataKeyPreviousHealth: s.PreviousHealth,
			MetadataKeyMessage:        s.Message,
			MetadataKeyComponentName:  s.ComponentName,
			MetadataKeyResourceKind:   s.RootResourceKind,
			MetadataKeyResourceNS:     s.RootResourceNamespace,
			MetadataKeyResourceName:   s.RootResourceName,
		},
	}
}

// ComponentUnhealthySignal fires when a component's verdict crosses into
// degraded/unhealthy. Never for unknown — that's runner-inactivity's job, not N alerts.
type ComponentUnhealthySignal struct {
	ComponentSignal
}

var (
	_ signal.Signal                     = (*ComponentUnhealthySignal)(nil)
	_ signal.SignalWithLifecycleContext = (*ComponentUnhealthySignal)(nil)
)

func (s *ComponentUnhealthySignal) Type() signal.SignalType { return ComponentUnhealthySignalType }

func (s *ComponentUnhealthySignal) LifecycleContext() signal.SignalLifecycleContext {
	return s.lifecycleContext("component-unhealthy")
}

func (s *ComponentUnhealthySignal) Validate(_ workflow.Context) error { return s.validate() }

func (s *ComponentUnhealthySignal) Execute(_ workflow.Context) error { return nil }

// ComponentRecoveredSignal fires when a component's verdict returns to
// healthy, so a channel told about a failure always gets its resolution.
type ComponentRecoveredSignal struct {
	ComponentSignal
}

var (
	_ signal.Signal                     = (*ComponentRecoveredSignal)(nil)
	_ signal.SignalWithLifecycleContext = (*ComponentRecoveredSignal)(nil)
)

func (s *ComponentRecoveredSignal) Type() signal.SignalType { return ComponentRecoveredSignalType }

func (s *ComponentRecoveredSignal) LifecycleContext() signal.SignalLifecycleContext {
	return s.lifecycleContext("component-recovered")
}

func (s *ComponentRecoveredSignal) Validate(_ workflow.Context) error { return s.validate() }

func (s *ComponentRecoveredSignal) Execute(_ workflow.Context) error { return nil }

// InstallDegradedSignal fires when the install's composite health crosses
// degraded/unhealthy; pairing it with per-component signals double-alerts one outage.
type InstallDegradedSignal struct {
	InstallID   string `json:"install_id"`
	InstallName string `json:"install_name"`

	Health         string `json:"health"`
	PreviousHealth string `json:"previous_health"`
	Message        string `json:"message"`

	UnhealthyComponentCount int `json:"unhealthy_component_count"`
	DegradedComponentCount  int `json:"degraded_component_count"`
}

var (
	_ signal.Signal                     = (*InstallDegradedSignal)(nil)
	_ signal.SignalWithLifecycleContext = (*InstallDegradedSignal)(nil)
)

func (s *InstallDegradedSignal) Type() signal.SignalType { return InstallDegradedSignalType }

func (s *InstallDegradedSignal) LifecycleContext() signal.SignalLifecycleContext {
	installID := &s.InstallID
	return signal.SignalLifecycleContext{
		InstallID: installID,
		Operation: "install-degraded",
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		OwnerName: s.InstallName,
		Metadata: map[string]any{
			MetadataKeyHealth:         s.Health,
			MetadataKeyPreviousHealth: s.PreviousHealth,
			MetadataKeyMessage:        s.Message,
			MetadataKeyUnhealthyCount: s.UnhealthyComponentCount,
			MetadataKeyDegradedCount:  s.DegradedComponentCount,
		},
	}
}

func (s *InstallDegradedSignal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.Health == "" {
		return errors.New("health is required")
	}
	return nil
}

func (s *InstallDegradedSignal) Execute(_ workflow.Context) error { return nil }
