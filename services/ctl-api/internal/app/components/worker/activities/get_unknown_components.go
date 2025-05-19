package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetUnknownComponents struct {
	Limit int
}

// @temporal-gen activity
func (a *Activities) GetUnknownComponents(ctx context.Context, req GetUnknownComponents) ([]app.Component, error) {
	comps, err := a.getUnkownComponents(ctx, req.Limit)
	if err != nil {
		return nil, err
	}

	return comps, nil
}

func (a *Activities) getUnkownComponents(ctx context.Context, limit int) ([]app.Component, error) {
	comps := make([]app.Component, 0)
	res := a.db.WithContext(ctx).Model(&app.Component{}).Where("type IS NULL").Order("created_at asc").Limit(limit).Find(&comps)
	if res.Error != nil {
		return nil, res.Error
	}

	return comps, nil
}
