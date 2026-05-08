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
// Callers represent "match everything" as a *nil* Selector, never as
// Selector{MatchLabels: nil} — see Validate.
type Selector struct {
	MatchLabels Labels `json:"match_labels"`
}

// Matches returns true iff every entry in MatchLabels is satisfied by set.
// Empty MatchLabels matches everything; callers should not construct such a
// Selector (Validate rejects it) but the matcher is permissive so dispatch
// never silently drops events on stale data.
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
	return true
}

// Validate enforces the invariant that a *non-nil* Selector carries at least
// one constraint. The "match everything" case is expressed by a nil pointer at
// the call site (TargetMatch.Selector == nil), not by an empty MatchLabels
// map — this keeps the dispatch / preview / canonical-key code paths from
// having to special-case an ambiguous shape.
func (s *Selector) Validate() error {
	if s == nil {
		return errors.New("selector is nil")
	}
	if len(s.MatchLabels) == 0 {
		return errors.New("selector match_labels must be non-empty (use a nil *Selector for match-all)")
	}
	for k := range s.MatchLabels {
		if strings.TrimSpace(k) == "" {
			return errors.New("selector match_labels has blank key")
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
	keys := make([]string, 0, len(s.MatchLabels))
	for k := range s.MatchLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([][2]string, len(keys))
	for i, k := range keys {
		pairs[i] = [2]string{k, s.MatchLabels[k]}
	}
	b, _ := json.Marshal(struct {
		MatchLabels [][2]string `json:"match_labels"`
	}{MatchLabels: pairs})
	return string(b)
}
