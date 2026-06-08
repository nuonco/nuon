package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetLatestActiveComponentBuildRequest struct {
	ComponentID string `validate:"required"`
}

// GetLatestActiveComponentBuild returns the most recent ComponentBuild for the
// given component whose status is Active. It is used by the install workflow
// generator to decide whether a non-image component's image dependencies have
// a newer Active build than what is currently deployed, so a sync step can be
// prepended for dep-aware deploys.
//
// Returns (nil, nil) when no Active build exists for the component yet.
//
// @temporal-gen-v2 activity
// @by-field ComponentID
func (a *Activities) GetLatestActiveComponentBuild(ctx context.Context, req GetLatestActiveComponentBuildRequest) (*app.ComponentBuild, error) {
	var build app.ComponentBuild
	viewOrTable := views.TableOrViewName(a.db, &app.ComponentConfigConnection{}, "")
	res := a.db.WithContext(ctx).
		Joins(fmt.Sprintf("JOIN %s ON %s.id=component_builds.component_config_connection_id", viewOrTable, viewOrTable)).
		Joins(fmt.Sprintf("JOIN components ON components.id=%s.component_id", viewOrTable)).
		Where("components.id = ?", req.ComponentID).
		Where("component_builds.status = ?", app.ComponentBuildStatusActive).
		Order("component_builds.created_at DESC").
		First(&build)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to load latest active component build: %w", res.Error)
	}

	return &build, nil
}
