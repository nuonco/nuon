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
// enabled. Mirror of the orgs-namespace activity of the same name; needed
// because workflows running in the components namespace cannot dispatch
// activities to the orgs worker.
//
// @temporal-gen-v2 activity
func (a *Activities) OrgHasFeature(ctx context.Context, req OrgHasFeatureRequest) (bool, error) {
	return a.features.OrgHasFeature(ctx, req.OrgID, app.OrgFeature(req.Feature))
}
