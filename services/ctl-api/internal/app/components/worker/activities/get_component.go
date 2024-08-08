package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentRequest struct {
	ComponentID string `validate:"required"`
}

// @await-gen
// @execution-timeout 5s
func (a *Activities) GetComponent(ctx context.Context, req GetComponentRequest) (*app.Component, error) {
	cmp := app.Component{}
	res := a.db.WithContext(ctx).
		Preload("ComponentConfigs").
		Preload("ComponentConfigs.ComponentBuilds").
		Preload("ComponentConfigs.ComponentBuilds.ComponentReleases").
		First(&cmp, "id = ?", req.ComponentID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &cmp, nil
}
