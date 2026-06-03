package labels

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// Selector matches a label set with AND across keys. A value of "*" means
// "key must exist" (mirrors the wildcard semantics of WithLabels' jsonb_exists
// branch in scopes.go); any other value requires exact equality (mirrors @>).
//
// NotMatchLabels expresses negation with the same value semantics: an entry
// rejects sets where the key is present and matches (exact value or wildcard
// existence). It exists so users can say "everything except env=stage" without
// having to enumerate every other env value or label every install.
//
// Callers represent "match everything" as a *nil* Selector, never as
// Selector{MatchLabels: nil} — see Validate.
type Selector struct {
	MatchLabels    Labels `json:"match_labels,omitempty"`
	NotMatchLabels Labels `json:"not_match_labels,omitempty"`
}

// Matches returns true iff every entry in MatchLabels is satisfied by set
// AND no entry in NotMatchLabels is satisfied by set. Empty MatchLabels and
// empty NotMatchLabels both mean "no constraint of that polarity"; callers
// should not construct such a Selector (Validate rejects it) but the matcher
// is permissive so dispatch never silently drops events on stale data.
func (s *Selector) Matches(set Labels) bool {
	if s == nil {
		return true
	}
	for k, v := range s.MatchLabels {
		got, ok := set[k]
		if !ok {
			return false
		}
		if v == "*" {
			continue
		}
		if got != v {
			return false
		}
	}
	for k, v := range s.NotMatchLabels {
		got, ok := set[k]
		if !ok {
			// key absent → exclusion not triggered for this entry
			continue
		}
		if v == "*" {
			// key-existence exclusion: present → reject
			return false
		}
		if got == v {
			return false
		}
	}
	return true
}

// Validate enforces the invariant that a *non-nil* Selector carries at least
// one constraint (positive or negative). The "match everything" case is
// expressed by a nil pointer at the call site (TargetMatch.Selector == nil),
// not by empty maps — this keeps the dispatch / preview / canonical-key code
// paths from having to special-case an ambiguous shape.
func (s *Selector) Validate() error {
	if s == nil {
		return errors.New("selector is nil")
	}
	if len(s.MatchLabels) == 0 && len(s.NotMatchLabels) == 0 {
		return errors.New("selector must have at least one of match_labels or not_match_labels (use a nil *Selector for match-all)")
	}
	for k := range s.MatchLabels {
		if strings.TrimSpace(k) == "" {
			return errors.New("selector match_labels has blank key")
		}
	}
	for k := range s.NotMatchLabels {
		if strings.TrimSpace(k) == "" {
			return errors.New("selector not_match_labels has blank key")
		}
	}
	return nil
}

// Canonical returns a deterministic JSON encoding (keys sorted) so the value
// can participate in the SlackChannelSubscription unique index without being
// at the mercy of Go map iteration order.
func (s *Selector) Canonical() string {
	if s == nil {
		return ""
	}
	sortedPairs := func(m Labels) [][2]string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pairs := make([][2]string, len(keys))
		for i, k := range keys {
			pairs[i] = [2]string{k, m[k]}
		}
		return pairs
	}
	b, _ := json.Marshal(struct {
		MatchLabels    [][2]string `json:"match_labels"`
		NotMatchLabels [][2]string `json:"not_match_labels,omitempty"`
	}{
		MatchLabels:    sortedPairs(s.MatchLabels),
		NotMatchLabels: sortedPairs(s.NotMatchLabels),
	})
	return string(b)
}
