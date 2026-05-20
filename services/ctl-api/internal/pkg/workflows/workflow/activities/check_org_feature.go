package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CheckOrgFeatureRequest struct {
	Feature string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field Feature
func (a *Activities) CheckOrgFeature(ctx context.Context, req CheckOrgFeatureRequest) (bool, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "unable to get org id from context")
	}

	var org app.Org
	if res := a.db.WithContext(ctx).First(&org, "id = ?", orgID); res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to fetch org")
	}

	val, ok := org.Features[string(app.OrgFeature(req.Feature))]
	if !ok {
		return false, nil
	}
	return val, nil
}
