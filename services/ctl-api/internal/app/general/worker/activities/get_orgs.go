package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetOrgsRequest struct{}

// @temporal-gen activity
func (a *Activities) GetOrgs(ctx context.Context, req GetOrgsRequest) ([]*app.Org, error) {
	var orgs []*app.Org

	res := a.db.WithContext(ctx).
		Order("priority desc").
		Select("id").
		Find(&orgs)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get orgs")
	}

	return orgs, nil
}
