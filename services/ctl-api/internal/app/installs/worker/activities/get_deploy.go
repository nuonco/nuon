package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetDeployRequest struct {
	DeployID string `validate:"required"`
}

func (a *Activities) GetDeploy(ctx context.Context, req GetDeployRequest) (*app.InstallDeploy, error) {
	installDeploy := app.InstallDeploy{}
	res := a.db.WithContext(ctx).
		Preload("ComponentBuild").
		First(&installDeploy, "id = ?", req.DeployID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	return &installDeploy, nil
}
