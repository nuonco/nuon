package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateDeployForRunbookRequest struct {
	InstallID   string `validate:"required"`
	ComponentID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateDeployForRunbook(ctx context.Context, req CreateDeployForRunbookRequest) (*app.InstallDeploy, error) {
	var installComponent app.InstallComponent
	res := a.db.WithContext(ctx).
		Where(app.InstallComponent{
			InstallID:   req.InstallID,
			ComponentID: req.ComponentID,
		}).
		First(&installComponent)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find install component: %w", res.Error)
	}

	// Get latest build for this component
	var build app.ComponentBuild
	res = a.db.WithContext(ctx).
		Where(app.ComponentBuild{ComponentID: req.ComponentID}).
		Order("created_at DESC").
		First(&build)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find latest build for component: %w", res.Error)
	}

	deploy := app.InstallDeploy{
		InstallComponentID: installComponent.ID,
		ComponentBuildID:   build.ID,
		ComponentID:        req.ComponentID,
		Type:               app.InstallDeployTypeApply,
	}

	res = a.db.WithContext(ctx).Create(&deploy)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create deploy: %w", res.Error)
	}

	return &deploy, nil
}
