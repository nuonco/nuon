package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type HasFeatureRequest struct {
	Feature string `validate:"required"`
}

// @temporal-gen activity
// @by-id Feature
func (a *Activities) HasFeature(ctx context.Context, req HasFeatureRequest) (bool, error) {
	return a.features.FeatureEnabled(ctx, app.OrgFeature(req.Feature))
}
