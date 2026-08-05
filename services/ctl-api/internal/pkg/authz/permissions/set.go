package permissions

import "strings"

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

func (p Set) CanPerform(obj string, perm Permission) error {
	objects := []string{obj, "*"}
	if parent, _, found := strings.Cut(obj, ":"); found {
		objects = []string{obj, parent, "*"}
	}

	for _, object := range objects {
		if val, ok := p[object]; ok && (val == PermissionAll || val == perm) {
			return nil
		}
	}

	return NoAccessError{
		Permission: perm,
		ObjectID:   obj,
	}
}
