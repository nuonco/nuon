// Package interests defines the shared "what events do you want?" model used
// by both Slack channel subscriptions and webhooks.
//
// One JSONB shape is stored on both subscriber tables. One classifier turns a
// signal lifecycle event into a list of slugs (resource:..., op:..., outcome:...,
// event:...) and one matcher decides whether a given Interests config wants to
// receive that event. UI variants (slack vs webhook) collapse/expand the same
// underlying fields differently but persist identically.
//
// AllEvents is the new-subscription default; once a user opts into per-resource
// configuration the Resources map becomes the source of truth and AllEvents
// must be false.
package interests

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// ResourceKind is the top-level event taxonomy exposed in the picker UI. The
// classifier maps each lifecycle / approval event to exactly one ResourceKind.
type ResourceKind string

const (
	ResourceInstalls              ResourceKind = "installs"
	ResourceComponents            ResourceKind = "components"
	ResourceSandboxes             ResourceKind = "sandboxes"
	ResourceInstallConfigurations ResourceKind = "install_configurations"
	ResourceRunners               ResourceKind = "runners"
	ResourceActions               ResourceKind = "actions"
)

// AllResources is the canonical, ordered list of resource kinds. UI code
// renders rows in this order; defaults() / docs walk it.
var AllResources = []ResourceKind{
	ResourceInstalls,
	ResourceComponents,
	ResourceSandboxes,
	ResourceInstallConfigurations,
	ResourceRunners,
	ResourceActions,
}

// Outcome filters lifecycle events on terminal status. Approval events are
// gated separately by ResourceCfg.ApprovalRequests / ApprovalResponses and
// ignore this field.
type Outcome string

const (
	// OutcomeNone mutes lifecycle events for the resource entirely. Approval and
	// drift-detected events on the same resource are unaffected — they are gated
	// independently by ApprovalRequests / ApprovalResponses / DriftDetected.
	// Use this with DriftDetected=true to subscribe to drift-only, or with
	// ApprovalRequests / ApprovalResponses to subscribe to approval-only flows.
	OutcomeNone Outcome = "none"
	// OutcomeAll forwards every started + terminal lifecycle event.
	OutcomeAll Outcome = "all"
	// OutcomeCompletion forwards only terminal events (succeeded / failed /
	// cancelled). Started events are suppressed.
	OutcomeCompletion Outcome = "completion"
	// OutcomeFailures forwards only failed / cancelled terminal events.
	OutcomeFailures Outcome = "failures"
)

// SubOps is the canonical sub-op vocabulary per resource. The classifier maps
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
	ResourceInstalls:              {"provision", "deprovision", "reprovision"},
	ResourceComponents:            {"deploy", "teardown"},
	ResourceSandboxes:             {"provision", "reprovision", "deprovision"},
	ResourceInstallConfigurations: {"inputs", "secrets"},
	ResourceRunners:               {"provision", "reprovision", "inactive"},
	ResourceActions:               {"run"},
}

// SupportsDriftDetected reports whether DriftDetected is meaningful for the
// given resource. Currently true for components and sandboxes — the only
// resources whose workflows can produce a drift-detected event.
func SupportsDriftDetected(kind ResourceKind) bool {
	return kind == ResourceComponents || kind == ResourceSandboxes
}

// Interests is the full per-subscriber config. Stored as JSONB on both
// slack_channel_subscriptions and webhooks.
//
// AllEvents acts as a sentinel: when true, every supported event matches
// regardless of Resources content. Toggling AllEvents off in the picker
// materializes Default() (the per-resource opt-out baseline) into Resources
// and flips AllEvents to false.
type Interests struct {
	AllEvents bool                         `json:"all_events,omitempty"`
	Resources map[ResourceKind]ResourceCfg `json:"resources,omitempty"`
}

// ResourceCfg is the per-resource filter. An empty Ops slice means "every
// sub-op for this resource". Outcome=="" is treated as OutcomeAll.
//
// ApprovalRequests / ApprovalResponses are independent of Outcome — approval
// events do not have a started/succeeded/failed lifecycle, only a requested /
// approved / rejected handshake.
//
// DriftDetected is independent of Outcome too. It gates the drift_detected
// event that fires from inside the plan-only check of a drift_run /
// drift_run_reprovision_sandbox workflow when the plan has changes. Only
// meaningful for resources where SupportsDriftDetected returns true
// (components, sandboxes); set on other resources is harmless but never
// matches. Drift workflow lifecycle events themselves are unconditionally
// suppressed by the matcher — DriftDetected is the only knob that surfaces
// drift to subscribers.
//
// LOCKSTEP: changes to this struct's JSON shape must also update
// bins/cli/cmd/orgs.go (interestsJSONHelp) and the "Filtering events with
// interests" section of docs/guides/webhooks.mdx.
type ResourceCfg struct {
	Ops               []string `json:"ops,omitempty"`
	Outcome           Outcome  `json:"outcome,omitempty"`
	ApprovalRequests  bool     `json:"approval_requests,omitempty"`
	ApprovalResponses bool     `json:"approval_responses,omitempty"`
	DriftDetected     bool     `json:"drift_detected,omitempty"`
}

// IsZero is true for the zero-value Interests (no AllEvents, no resources).
// Used by callers to distinguish "never configured" from "explicitly empty".
func (i Interests) IsZero() bool {
	return !i.AllEvents && len(i.Resources) == 0
}

// Scan implements sql.Scanner so Interests can be loaded from a JSONB column.
// Mirrors the pattern used by IntermediateAppConfig in
// services/ctl-api/internal/app/onboarding.go.
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

// Value implements driver.Valuer so Interests can be persisted to a JSONB
// column.
func (i Interests) Value() (driver.Value, error) {
	if i.IsZero() {
		// Persist as JSON `null` so the column round-trips cleanly even when
		// neither AllEvents nor Resources are populated.
		return nil, nil
	}
	return json.Marshal(i)
}

// GormDataType tells GORM to materialise the column as JSONB.
func (Interests) GormDataType() string {
	return "jsonb"
}
