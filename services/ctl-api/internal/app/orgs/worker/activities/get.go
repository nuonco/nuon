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
	org := app.Org{}
	res := a.db.WithContext(ctx).
		Preload("Apps").
		First(&org, "id = ?", req.OrgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &org, nil
}
