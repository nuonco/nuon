package activities

import "context"

type GCPLegacyRoleFallbackEnabledRequest struct{}

// @temporal-gen-v2 activity
func (a *Activities) GCPLegacyRoleFallbackEnabled(ctx context.Context, req GCPLegacyRoleFallbackEnabledRequest) (bool, error) {
	return a.cfg.GCPLegacyRoleFallback, nil
}
