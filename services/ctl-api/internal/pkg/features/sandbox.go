package features

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func (f *Features) OrgType(ctx context.Context) (app.OrgType, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return app.OrgTypeUnknown, errors.Wrap(err, "unable to get org id")
	}

	var org app.Org
	if res := f.db.WithContext(ctx).
		First(&org, "id = ?", orgID); res.Error != nil {
		return app.OrgTypeUnknown, errors.Wrap(res.Error, "unable to fetch org")
	}

	return org.OrgType, nil
}
