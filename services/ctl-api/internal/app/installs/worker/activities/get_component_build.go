package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentBuildRequest struct {
	ComponentBuildID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentBuildID
func (a *Activities) GetComponentBuild(ctx context.Context, req GetComponentBuildRequest) (*app.ComponentBuild, error) {
	var build app.ComponentBuild
	res := a.db.WithContext(ctx).
		Where("id = ?", req.ComponentBuildID).
		First(&build)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component build: %w", res.Error)
	}

	return &build, nil
}
