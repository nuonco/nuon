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

		p[k] = perm
	}

	return nil
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
