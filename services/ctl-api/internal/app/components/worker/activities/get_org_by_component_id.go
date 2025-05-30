package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetOrgRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetOrg(ctx context.Context, req GetOrgRequest) (*app.Org, error) {
	var org app.Org
	if res := a.db.WithContext(ctx).First(&org, "id = ?", req.ID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get org")
	}

	return &org, nil
}
