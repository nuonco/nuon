package authz

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

// Link is one tier in a resource's ownership chain: its object id and the grant
// resource type of that tier (e.g. {"install", "inl_..."}). The type lets
// wildcard grants (all resources of a type) authorize without leaking to
// other tiers.
type Link struct {
	Type string
	ID   string
}

// Authorize walks a resource's ownership chain (most-specific first, e.g.
// [install, app, org]) and allows the request if any link carries a permission
// that satisfies perm — either a grant on that specific object (perms) or a
// wildcard grant on that link's tier (wildcards, keyed by resource type). This
// unifies object-level grants with the existing org-wide check: the org id is
// simply the last link in every chain.
func Authorize(perms permissions.Set, wildcards map[string]permissions.Permission, chain []Link, perm permissions.Permission) error {
	var lastErr error
	for _, link := range chain {
		if link.ID == "" {
			continue
		}
		if err := perms.CanPerform(link.ID, perm); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if w, ok := wildcards[link.Type]; ok && (w == permissions.PermissionAll || w == perm) {
			return nil
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return permissions.NoAccessError{Permission: perm}
}
