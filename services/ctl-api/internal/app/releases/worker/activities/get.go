package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	ReleaseID string `validate:"required"`
}

func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.ComponentRelease, error) {
	release := app.ComponentRelease{}
	res := a.db.WithContext(ctx).
		Preload("ComponentBuild").
		Preload("ComponentReleaseSteps").
		First(&release, "id = ?", req.ReleaseID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get release: %w", res.Error)
	}

	return &release, nil
}
