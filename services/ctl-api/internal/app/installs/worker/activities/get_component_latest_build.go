package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentLatestBuildRequest struct {
	ComponentID string `validate:"required"`
}

func (a *Activities) GetComponentLatestBuild(ctx context.Context, req GetComponentLatestBuildRequest) (*app.ComponentBuild, error) {
	var build app.ComponentBuild
	res := a.db.WithContext(ctx).
		Joins("JOIN component_config_connections_view ON component_config_connections_view.id=component_builds.component_config_connection_id").
		Joins("JOIN components ON components.id=component_config_connections_view.component_id").
		Where("components.id = ?", req.ComponentID).
		Order("created_at DESC").
		First(&build)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component build: %w", res.Error)
	}

	return &build, nil
}
