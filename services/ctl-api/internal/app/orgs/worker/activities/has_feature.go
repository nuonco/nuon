package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type OrgHasFeatureRequest struct {
	OrgID   string `validate:"required"`
	Feature string `validate:"required"`
}

// OrgHasFeature returns true if the given org has the given feature flag
// enabled. Useful from workflows that have an OrgID in hand but no
// org-context-aware ctx (so callers cannot rely on
// features.FeatureEnabled, which reads OrgID from cctx).
//
// @temporal-gen-v2 activity
func (a *Activities) OrgHasFeature(ctx context.Context, req OrgHasFeatureRequest) (bool, error) {
	return a.features.OrgHasFeature(ctx, req.OrgID, app.OrgFeature(req.Feature))
}
