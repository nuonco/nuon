package permissions

import (
	"encoding/json"
	"slices"
	"strings"
)

// Verbs is the set of verbs granted on a single object. PermissionAll
// subsumes the other verbs, so a Verbs containing it is normalized to
// exactly [all].
type Verbs []Permission

func (v Verbs) Can(perm Permission) bool {
	return slices.Contains(v, PermissionAll) || slices.Contains(v, perm)
}

// With returns the union of v and perms. Granting is additive across an
// account's roles, so the same object appearing in multiple policies merges
// to the stronger set regardless of iteration order.
func (v Verbs) With(perms ...Permission) Verbs {
	if slices.Contains(v, PermissionAll) {
		return Verbs{PermissionAll}
	}

	out := slices.Clone(v)
	for _, perm := range perms {
		if perm == PermissionAll {
			return Verbs{PermissionAll}
		}
		if !slices.Contains(out, perm) {
			out = append(out, perm)
		}
	}
	return out
}

// Set maps an object id to the verbs granted on it.
//
// On the wire a Set stays a map of single verb strings, the shape it had
// before verb sets existed, so already-released clients keep parsing it (see
// MarshalJSON). Fields of this type are tagged swaggertype:"object,string" to
// match.
type Set map[string]Verbs

func NewSet() Set {
	return make(Set, 0)
}

// legacyOrder lists verbs least-privilege first, so a set that cannot be
// expressed as one verb reports the weakest one it holds rather than
// overstating the grant.
var legacyOrder = []Permission{PermissionRead, PermissionCreate, PermissionUpdate, PermissionDelete}

// MarshalJSON emits the pre-verb-set shape: one verb string per object.
// Verb subsets are not representable there, so a subset reports its
// least-privileged verb; only a complete set reports "all". Nothing reads this
// field to make authorization decisions — the server authorizes from the Go
// value, and scoped detail lives on Policy.ScopedPermissions — so the
// lossiness buys wire compatibility at no cost.
func (p Set) MarshalJSON() ([]byte, error) {
	out := make(map[string]Permission, len(p))
	for obj, verbs := range p {
		out[obj] = verbs.legacyVerb()
	}
	return json.Marshal(out)
}

// UnmarshalJSON accepts both the legacy single-verb shape and a verb array, so
// a Set survives a round trip through either representation.
func (p *Set) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	out := make(Set, len(raw))
	for obj, val := range raw {
		var verbs Verbs
		if err := json.Unmarshal(val, &verbs); err != nil {
			var single Permission
			if err := json.Unmarshal(val, &single); err != nil {
				return err
			}
			verbs = Verbs{single}
		}
		out[obj] = verbs
	}

	*p = out
	return nil
}

func (v Verbs) legacyVerb() Permission {
	if len(v) == 0 {
		return PermissionUnknown
	}
	if v.Can(PermissionAll) {
		return PermissionAll
	}

	complete := true
	for _, perm := range legacyOrder {
		if !slices.Contains(v, perm) {
			complete = false
			break
		}
	}
	if complete {
		return PermissionAll
	}

	for _, perm := range legacyOrder {
		if slices.Contains(v, perm) {
			return perm
		}
	}
	return PermissionUnknown
}

func (p Set) Add(set map[string]*string) error {
	for k, v := range set {
		perm, err := NewPermission(*v)
		if err != nil {
			return err
		}

		p[k] = p[k].With(perm)
	}

	return nil
}

// Grant adds verbs on a single object, merging with any existing grant.
func (p Set) Grant(obj string, verbs ...Permission) {
	p[obj] = p[obj].With(verbs...)
}

func (p Set) CanPerform(obj string, perm Permission) error {
	objects := []string{obj, "*"}
	if parent, _, found := strings.Cut(obj, ":"); found {
		objects = []string{obj, parent, "*"}
	}

	for _, object := range objects {
		if p[object].Can(perm) {
			return nil
		}
	}

	return NoAccessError{
		Permission: perm,
		ObjectID:   obj,
	}
}
