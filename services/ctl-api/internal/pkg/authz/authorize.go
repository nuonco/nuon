package authz

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

// Authorize walks a resource's ownership chain (most-specific first, e.g.
// [installID, appID, orgID]) and allows the request if any link carries a
// permission that satisfies perm. This unifies object-level grants with the
// existing org-wide check — the org id is simply the last link in every chain.
func Authorize(perms permissions.Set, chain []string, perm permissions.Permission) error {
	var lastErr error
	for _, id := range chain {
		if id == "" {
			continue
		}
		if err := perms.CanPerform(id, perm); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return permissions.NoAccessError{Permission: perm}
}
