package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetOrgInstallsRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen activity
// @by-id OrgID
func (a *Activities) GetOrgInstalls(ctx context.Context, req GetOrgInstallsRequest) ([]*app.Install, error) {
	var installs []*app.Install

	res := a.db.WithContext(ctx).
		Where(app.Install{
			OrgID: req.OrgID,
		}).
		Select("id").
		Find(&installs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get org installs")
	}

	return installs, nil
}
