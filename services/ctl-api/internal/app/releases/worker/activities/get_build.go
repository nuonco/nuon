package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetBuildRequest struct {
	BuildID string `validate:"required"`
}

func (a *Activities) GetBuild(ctx context.Context, req GetBuildRequest) (*app.ComponentBuild, error) {
	build := app.ComponentBuild{}
	res := a.db.WithContext(ctx).
		First(&build, "id = ?", req.BuildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get build: %w", res.Error)
	}

	return &build, nil
}
