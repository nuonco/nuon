package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type OrgHasFeatureRequest struct {
	OrgID   string `json:"org_id" validate:"required"`
	Feature string `json:"feature" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) OrgHasFeature(ctx context.Context, req OrgHasFeatureRequest) (bool, error) {
	return a.features.OrgHasFeature(ctx, req.OrgID, app.OrgFeature(req.Feature))
}
