package labels

import (
	"database/sql/driver"
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
)

type TargetKind string

const (
	TargetKindInstalls    TargetKind = "installs"
	TargetKindComponents  TargetKind = "components"
	TargetKindActions     TargetKind = "actions"
	TargetKindAppBranches TargetKind = "app_branches"
)

type EventTargets struct {
	InstallID       string
	ComponentID     string
	ActionID        string
	AppBranchID     string
	InstallLabels   Labels
	ComponentLabels Labels
	ActionLabels    Labels
}

type TargetMatch struct {
	IDs      []string  `json:"ids,omitempty"`
	Selector *Selector `json:"selector,omitempty"`
}

// why: matches reports whether (id, lbls) satisfies this filter. id == "" is
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

type SubscriptionMatch struct {
	Installs    *TargetMatch `json:"installs,omitempty"`
	Components  *TargetMatch `json:"components,omitempty"`
	Actions     *TargetMatch `json:"actions,omitempty"`
	AppBranches *TargetMatch `json:"app_branches,omitempty"`
}

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
	if m.AppBranches != nil && m.AppBranches.matches(t.AppBranchID, nil) {
		return true
	}
	return false
}

func (m *SubscriptionMatch) isZero() bool {
	if m == nil {
		return true
	}
	return m.Installs == nil && m.Components == nil && m.Actions == nil && m.AppBranches == nil
}

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
	if m.AppBranches != nil {
		if err := m.AppBranches.Validate(); err != nil {
			return errors.Wrap(err, "app_branches")
		}
	}
	return nil
}

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
		Installs    *tmCanon `json:"installs,omitempty"`
		Components  *tmCanon `json:"components,omitempty"`
		Actions     *tmCanon `json:"actions,omitempty"`
		AppBranches *tmCanon `json:"app_branches,omitempty"`
	}{
		Installs:    canon(m.Installs),
		Components:  canon(m.Components),
		Actions:     canon(m.Actions),
		AppBranches: canon(m.AppBranches),
	})
	return string(b)
}

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

func (m SubscriptionMatch) Value() (driver.Value, error) {
	if m.isZero() {
		return nil, nil
	}
	return json.Marshal(m)
}

func (SubscriptionMatch) GormDataType() string {
	return "jsonb"
}
