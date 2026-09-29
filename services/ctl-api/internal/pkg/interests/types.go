package interests

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type ResourceKind string

const (
	ResourceInstalls              ResourceKind = "installs"
	ResourceStacks                ResourceKind = "stacks"
	ResourceComponents            ResourceKind = "components"
	ResourceSandboxes             ResourceKind = "sandboxes"
	ResourceInstallConfigurations ResourceKind = "install_configurations"
	ResourceRunners               ResourceKind = "runners"
	ResourceActions               ResourceKind = "actions"
	ResourceAppBranches           ResourceKind = "app_branches"
)

var AllResources = []ResourceKind{
	ResourceInstalls,
	ResourceStacks,
	ResourceComponents,
	ResourceSandboxes,
	ResourceInstallConfigurations,
	ResourceRunners,
	ResourceActions,
	ResourceAppBranches,
}

type Outcome string

const (
	OutcomeNone       Outcome = "none"
	OutcomeAll        Outcome = "all"
	OutcomeCompletion Outcome = "completion"
	OutcomeFailures   Outcome = "failures"
)

// why: SubOps is the canonical sub-op vocabulary per resource. The classifier maps
// real WorkflowType constants from internal/app/workflow.go onto these slugs.
// UIs render checkbox lists from this map.
//
// Note: "drift" is intentionally NOT listed for components or sandboxes even
// though the classifier still emits the slug. Drift workflow lifecycle events
// are pure noise (one started/completed pair per cron tick per resource) and
// are unconditionally suppressed in match.go. Subscribers opt into drift
// notifications through DriftDetected, which gates the dedicated drift-detected
// event that only fires when the plan-only check observes actual changes.
var SubOps = map[ResourceKind][]string{
	ResourceInstalls:              {"provision", "deprovision", "reprovision", "label_added", "app_branch_changed"},
	ResourceStacks:                {"version_active", "stack_run", "role_change", "inputs_updated"},
	ResourceComponents:            {"deploy", "teardown"},
	ResourceSandboxes:             {"provision", "reprovision", "deprovision"},
	ResourceInstallConfigurations: {"inputs", "secrets", "sync"},
	ResourceRunners:               {"provision", "reprovision", "inactive", "unhealthy"},
	ResourceActions:               {"run"},
	ResourceAppBranches:           {"run"},
}

func SupportsDriftDetected(kind ResourceKind) bool {
	return kind == ResourceComponents || kind == ResourceSandboxes
}

func SupportsRoleChanges(kind ResourceKind) bool {
	return kind == ResourceStacks
}

func SupportsInputsUpdated(kind ResourceKind) bool {
	return kind == ResourceStacks
}

func SupportsConfigSynced(kind ResourceKind) bool {
	return kind == ResourceAppBranches
}

func SupportsComponentHealth(kind ResourceKind) bool {
	return kind == ResourceComponents
}

func SupportsInstallDegraded(kind ResourceKind) bool {
	return kind == ResourceInstalls
}

type Interests struct {
	AllEvents bool                         `json:"all_events,omitempty"`
	Resources map[ResourceKind]ResourceCfg `json:"resources,omitempty"`
}

type ResourceCfg struct {
	Ops               []string `json:"ops,omitempty"`
	Outcome           Outcome  `json:"outcome,omitempty"`
	ApprovalRequests  bool     `json:"approval_requests,omitempty"`
	ApprovalResponses bool     `json:"approval_responses,omitempty"`
	DriftDetected     bool     `json:"drift_detected,omitempty"`
	RoleChanges       bool     `json:"role_changes,omitempty"`
	InputsUpdated     bool     `json:"inputs_updated,omitempty"`
	ConfigSynced      bool     `json:"config_synced,omitempty"`
	ComponentHealth   bool     `json:"component_health,omitempty"`
	InstallDegraded   bool     `json:"install_degraded,omitempty"`
}

func (i Interests) IsZero() bool {
	return !i.AllEvents && len(i.Resources) == 0
}

func (i *Interests) Scan(v interface{}) error {
	switch v := v.(type) {
	case nil:
		*i = Interests{}
		return nil
	case []byte:
		if len(v) == 0 {
			*i = Interests{}
			return nil
		}
		if err := json.Unmarshal(v, i); err != nil {
			return errors.Wrap(err, "unable to scan interests config")
		}
		return nil
	case string:
		if v == "" {
			*i = Interests{}
			return nil
		}
		if err := json.Unmarshal([]byte(v), i); err != nil {
			return errors.Wrap(err, "unable to scan interests config")
		}
		return nil
	default:
		return errors.Errorf("unsupported scan type for interests config: %T", v)
	}
}

func (i Interests) Value() (driver.Value, error) {
	if i.IsZero() {
		return nil, nil
	}
	return json.Marshal(i)
}

func (Interests) GormDataType() string {
	return "jsonb"
}
