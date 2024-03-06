package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetOrgRequest struct {
	AppID string `validate:"required"`
}

func (a *Activities) GetOrg(ctx context.Context, req GetOrgRequest) (*app.Org, error) {
	org := app.Org{}
	res := a.db.WithContext(ctx).
		Joins("JOIN apps on apps.org_id = orgs.id").
		Where("apps.id = ?", req.AppID).
		First(&org)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &org, nil
}
