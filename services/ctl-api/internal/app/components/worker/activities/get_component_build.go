package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentBuildRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetComponentBuild(ctx context.Context, req GetComponentBuildRequest) (*app.ComponentBuild, error) {
	return a.getComponentBuild(ctx, req.ID)
}

func (a *Activities) getComponentBuild(ctx context.Context, buildID string) (*app.ComponentBuild, error) {
	bld := app.ComponentBuild{}
	res := a.db.WithContext(ctx).
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.Component").
		First(&bld, "id = ?", buildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component build: %w", res.Error)
	}

	return &bld, nil
}
