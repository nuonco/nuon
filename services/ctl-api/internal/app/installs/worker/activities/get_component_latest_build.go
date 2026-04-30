package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type GetComponentLatestBuildRequest struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ComponentID
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
			// Terminal: retrying will not produce a build. Step-level
			// auto-retry short-circuits on this marker (see
			// signal.IsTerminalError).
			return nil, signal.NewTerminalError(
				"no_component_build",
				"No active build found for component (id %s). Ensure there is an active build for the component before retrying.",
				req.ComponentID,
			)
		}

		return nil, fmt.Errorf("unable to load component build: %w", res.Error)
	}

	return &build, nil
}
