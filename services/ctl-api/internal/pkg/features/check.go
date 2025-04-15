package features

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func (f *Features) OrgHasFeature(ctx context.Context, orgID string, feature app.OrgFeature) (bool, error) {
	var org app.Org
	if res := f.db.WithContext(ctx).
		First(&org, "id = ?", orgID); res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to fetch org")
	}

	val, ok := org.Features[string(feature)]
	if !ok {
		return false, nil
	}

	return val, nil
}

func (f *Features) FeatureEnabled(ctx context.Context, feature app.OrgFeature) (bool, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return false, nil
	}

	var org app.Org
	if res := f.db.WithContext(ctx).
		First(&org, "id = ?", orgID); res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to fetch org")
	}

	val, ok := org.Features[string(feature)]
	if !ok {
		return false, nil
	}

	return val, nil
}
