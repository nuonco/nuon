package syncer

import (
	"context"
)

// syncAppBreakGlass creates the app break-glass configuration.
// Note: Break-glass roles are handled as part of permissions sync
// This is a no-op since break-glass roles are added to AppPermissionsConfig
func (s *syncer) syncAppBreakGlass(ctx context.Context) error {
	// Break-glass roles are handled in syncAppPermissions
	// This method exists to maintain compatibility with the sync steps
	return nil
}
