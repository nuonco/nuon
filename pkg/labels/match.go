// Package labels' match.go defines the SubscriptionMatch shape persisted as a
// nullable JSONB column on slack_channel_subscriptions. A nil column value
// (and equivalently a nil receiver here) means "every event in the org" —
// this collapses the legacy scope=org|install + InstallID *string design into
// a single column.
//
// SubscriptionMatch composes per-resource TargetMatch slots; TargetMatch
// composes a list of IDs with an optional Selector. Composition rules:
//   - SubscriptionMatch.Matches: OR across non-nil kinds.
//   - TargetMatch.matches:       OR within (id ∈ IDs OR Selector hit).
//   - Selector.Matches:          AND across MatchLabels.
//
// TargetKind is intentionally NOT shared with interests.ResourceKind:
// pkg/labels lives below internal/ in the import graph and the resource
// taxonomy is a property of the dispatch layer, not of label matching.
package labels

import (
	"database/sql/driver"
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
)

// TargetKind is the set of entity kinds a SubscriptionMatch can target.
// Kept deliberately narrow — only kinds that the picker UI exposes today.
type TargetKind string

const (
	TargetKindInstalls   TargetKind = "installs"
	TargetKindComponents TargetKind = "components"
	TargetKindActions    TargetKind = "actions"
)

// EventTargets is the projection of a dispatched event the matcher needs:
// the resolved entity IDs and their label sets at the time of dispatch. Any
// field may be empty — an absent ID short-circuits TargetMatch.matches so a
// component-only event never falsely satisfies an installs filter.
type EventTargets struct {
	InstallID       string
	ComponentID     string
	ActionID        string
	InstallLabels   Labels
	ComponentLabels Labels
	ActionLabels    Labels
}

// TargetMatch filters one entity kind. An empty TargetMatch{} (no IDs, no
// Selector) is intentionally valid and means "any entity of this kind" —
// the modal needs an explicit "Any installs" option distinct from "Specific
// installs" / "By labels" without forcing a sentinel value.
type TargetMatch struct {
	IDs      []string  `json:"ids,omitempty"`
	Selector *Selector `json:"selector,omitempty"`
}

// matches reports whether (id, lbls) satisfies this filter. id == "" is
// treated as "no entity of this kind on the event" and never matches, even
// when the filter is the permissive empty TargetMatch{} — otherwise a
// component-only event would satisfy an installs target.
func (t *TargetMatch) matches(id string, lbls Labels) bool {
	if t == nil {
		return false
	}
	if id == "" {
		return false
	}
	// Empty filter ⇒ any entity of this kind (provided id != "" above).
	if len(t.IDs) == 0 && t.Selector == nil {
		return true
	}
	for _, candidate := range t.IDs {
		if candidate == id {
			return true
		}
	}
	if t.Selector != nil && t.Selector.Matches(lbls) {
		return true
	}
	return false
}

// Validate ensures the filter is well-formed. Empty TargetMatch{} is allowed
// (see type doc); a non-nil Selector must itself be non-empty.
func (t *TargetMatch) Validate() error {
	if t == nil {
		return errors.New("target match is nil")
	}
	for i, id := range t.IDs {
		if id == "" {
			return errors.Errorf("target match ids[%d] is empty", i)
		}
	}
	if t.Selector != nil {
		if err := t.Selector.Validate(); err != nil {
			return errors.Wrap(err, "target match selector")
		}
	}
	return nil
}

// SubscriptionMatch is the per-subscription routing filter persisted as JSONB
// on slack_channel_subscriptions.match. Schema supports populating multiple
// kinds in one row (CLI uses this); the v1 modal will only ever write one
// kind at a time.
type SubscriptionMatch struct {
	Installs   *TargetMatch `json:"installs,omitempty"`
	Components *TargetMatch `json:"components,omitempty"`
	Actions    *TargetMatch `json:"actions,omitempty"`
}

// Matches returns true if any non-nil kind matches its corresponding event
// target. A nil receiver is the org-wide subscription — match everything.
func (m *SubscriptionMatch) Matches(t EventTargets) bool {
	if m == nil {
		return true
	}
	if m.Installs != nil && m.Installs.matches(t.InstallID, t.InstallLabels) {
		return true
	}
	if m.Components != nil && m.Components.matches(t.ComponentID, t.ComponentLabels) {
		return true
	}
	if m.Actions != nil && m.Actions.matches(t.ActionID, t.ActionLabels) {
		return true
	}
	return false
}

// isZero reports the "no kinds populated" shape so Value() can persist it as
// SQL NULL — keeps the column's "nil = org-wide" invariant intact whether the
// caller passes a nil pointer or a zero-valued struct.
func (m *SubscriptionMatch) isZero() bool {
	if m == nil {
		return true
	}
	return m.Installs == nil && m.Components == nil && m.Actions == nil
}

// Validate requires at least one populated kind and recurses into each. A
// fully-empty SubscriptionMatch should be expressed as a nil column, not as a
// row carrying an unconstrained struct.
func (m *SubscriptionMatch) Validate() error {
	if m == nil {
		return errors.New("subscription match is nil (use a nil column for org-wide)")
	}
	if m.isZero() {
		return errors.New("subscription match has no kinds populated (use a nil column for org-wide)")
	}
	if m.Installs != nil {
		if err := m.Installs.Validate(); err != nil {
			return errors.Wrap(err, "installs")
		}
	}
	if m.Components != nil {
		if err := m.Components.Validate(); err != nil {
			return errors.Wrap(err, "components")
		}
	}
	if m.Actions != nil {
		if err := m.Actions.Validate(); err != nil {
			return errors.Wrap(err, "actions")
		}
	}
	return nil
}

// Canonical returns a deterministic JSON encoding suitable for use in the
// slack_channel_subscriptions unique index column. Sorting nested slices /
// selector keys here means two semantically-equal matches produce identical
// bytes regardless of construction order or map iteration. "" for the
// match-all (nil / zero) case so it round-trips with the SQL NULL.
func (m *SubscriptionMatch) Canonical() string {
	if m.isZero() {
		return ""
	}
	type tmCanon struct {
		IDs      []string `json:"ids,omitempty"`
		Selector string   `json:"selector,omitempty"`
	}
	canon := func(t *TargetMatch) *tmCanon {
		if t == nil {
			return nil
		}
		ids := append([]string(nil), t.IDs...)
		sort.Strings(ids)
		out := &tmCanon{IDs: ids}
		if t.Selector != nil {
			out.Selector = t.Selector.Canonical()
		}
		return out
	}
	b, _ := json.Marshal(struct {
		Installs   *tmCanon `json:"installs,omitempty"`
		Components *tmCanon `json:"components,omitempty"`
		Actions    *tmCanon `json:"actions,omitempty"`
	}{
		Installs:   canon(m.Installs),
		Components: canon(m.Components),
		Actions:    canon(m.Actions),
	})
	return string(b)
}

// Scan implements sql.Scanner for the JSONB match column. Mirrors the pattern
// used by interests.Interests so a NULL or empty column round-trips to a
// zero-valued struct (which Value() then turns back into NULL).
func (m *SubscriptionMatch) Scan(v interface{}) error {
	switch v := v.(type) {
	case nil:
		*m = SubscriptionMatch{}
		return nil
	case []byte:
		if len(v) == 0 {
			*m = SubscriptionMatch{}
			return nil
		}
		if err := json.Unmarshal(v, m); err != nil {
			return errors.Wrap(err, "unable to scan subscription match")
		}
		return nil
	case string:
		if v == "" {
			*m = SubscriptionMatch{}
			return nil
		}
		if err := json.Unmarshal([]byte(v), m); err != nil {
			return errors.Wrap(err, "unable to scan subscription match")
		}
		return nil
	default:
		return errors.Errorf("unsupported scan type for subscription match: %T", v)
	}
}

// Value implements driver.Valuer. The zero match persists as SQL NULL — see
// isZero — so the column's "nil = match everything" semantics survive
// callers that hand us &SubscriptionMatch{} instead of nil.
func (m SubscriptionMatch) Value() (driver.Value, error) {
	if m.isZero() {
		return nil, nil
	}
	return json.Marshal(m)
}

// GormDataType marks the column as JSONB.
func (SubscriptionMatch) GormDataType() string {
	return "jsonb"
}
