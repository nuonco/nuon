package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"gorm.io/gorm"
)

type GetComponentLatestBuildRequest struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentID
func (a *Activities) GetComponentLatestBuild(ctx context.Context, req GetComponentLatestBuildRequest) (*app.ComponentBuild, error) {
	var build app.ComponentBuild
	viewOrTable := views.TableOrViewName(a.db, &app.ComponentConfigConnection{}, "")
	res := a.db.WithContext(ctx).
		Joins(fmt.Sprintf("JOIN %s ON %s.id=component_builds.component_config_connection_id", viewOrTable, viewOrTable)).
		Joins(fmt.Sprintf("JOIN components ON components.id=%s.component_id", viewOrTable)).
		Where("components.id = ?", req.ComponentID).
		Order("created_at DESC").
		First(&build)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, generics.TemporalGormError(gorm.ErrRecordNotFound, "component build not found")
		}

		return nil, fmt.Errorf("unable to load component build: %w", res.Error)
	}

	return &build, nil
}
