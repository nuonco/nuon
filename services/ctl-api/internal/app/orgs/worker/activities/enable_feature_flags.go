package activities

import (
	"context"
)

type EnableFeaturesRequest struct {
	OrgID string `json:"org_id" validate:"required"`
}

// @temporal-gen activity
// @by-id OrgID
func (a *Activities) EnableFeatures(ctx context.Context, req EnableFeaturesRequest) error {
	return a.features.Enable(ctx, req.OrgID, map[string]bool{"all": true})
}
