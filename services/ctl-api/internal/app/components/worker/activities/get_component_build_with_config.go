package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
)

type GetComponentBuildWithConfigRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetComponentBuildWithConfig(ctx context.Context, req GetComponentBuildWithConfigRequest) (*app.ComponentBuild, error) {
	return a.getComponentBuildWithConfig(ctx, req.ID)
}

func (a *Activities) getComponentBuildWithConfig(ctx context.Context, buildID string) (*app.ComponentBuild, error) {
	var bld app.ComponentBuild

	res := a.db.WithContext(ctx).
		Scopes(helpers.PreloadComponentBuildConfig).
		First(&bld, "id = ?", buildID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get component build")
	}

	return &bld, nil
}
