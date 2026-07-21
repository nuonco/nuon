package permissions

type Set map[string]Permission

func NewSet() map[string]Permission {
	return make(map[string]Permission, 0)
}

func (p Set) Add(set map[string]*string) error {
	for k, v := range set {
		perm, err := NewPermission(*v)
		if err != nil {
			return err
		}

		// Merging is additive across an account's roles, so keep the stronger
		// grant when the same object appears in multiple policies. Without this,
		// a read-only policy could clobber an "all" grant from another role
		// depending on iteration order.
		if existing, ok := p[k]; ok && existing == PermissionAll {
			continue
		}

		p[k] = perm
	}

	return nil
}

// Grant adds a single object permission, preferring the most permissive value
// on collision so a narrow grant can never downgrade a broader existing grant
// (e.g. a role policy of {orgID: all} is not clobbered by {orgID: read}).
func (p Set) Grant(obj string, perm Permission) {
	if existing, ok := p[obj]; ok {
		if existing == PermissionAll || perm != PermissionAll {
			return
		}
	}

	p[obj] = perm
}

func (p Set) CanPerform(obj string, perm Permission) error {
	val, ok := p[obj]

	// if the object is not in the permission set, look up the "*" wildcard.
	if !ok {
		val, ok = p["*"]
	}

	// if still not found, return an error
	if !ok {
		return NoAccessError{
			Permission: perm,
			ObjectID:   obj,
		}
	}

	if val == PermissionAll || val == perm {
		return nil
	}

	return NoAccessError{
		Permission: perm,
		ObjectID:   obj,
	}
}
