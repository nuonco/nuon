package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
)

type GetComponentsWithType struct {
	IDs []string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetComponentsWithType(ctx context.Context, req GetComponentsWithType) ([]app.Component, error) {
	comps := make([]app.Component, 0)
	res := a.db.WithContext(ctx).Model(&app.Component{}).Where("ID IN ?", req.IDs).
		Scopes(helpers.PreloadLatestConfig).
		Find(&comps)
	if res.Error != nil {
		return nil, res.Error
	}

	return comps, nil
}
