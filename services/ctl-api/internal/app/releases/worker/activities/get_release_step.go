package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetReleaseStepRequest struct {
	ReleaseStepID string `validate:"required"`
}

func (a *Activities) GetReleaseStep(ctx context.Context, req GetReleaseStepRequest) (*app.ComponentReleaseStep, error) {
	step := app.ComponentReleaseStep{}
	res := a.db.WithContext(ctx).
		Preload("ComponentRelease").
		Preload("ComponentRelease.ComponentBuild").
		First(&step, "id = ?", req.ReleaseStepID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get release step: %w", res.Error)
	}

	return &step, nil
}
