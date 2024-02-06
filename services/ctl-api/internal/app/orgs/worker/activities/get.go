package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	OrgID string `validate:"required"`
}

func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.Org, error) {
	org, err := a.getOrg(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	return org, nil
}

func (a *Activities) getOrg(ctx context.Context, orgID string) (*app.Org, error) {
	org := app.Org{}
	res := a.db.WithContext(ctx).
		Preload("Apps").
		Preload("Apps.Installs").
		First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &org, nil
}
