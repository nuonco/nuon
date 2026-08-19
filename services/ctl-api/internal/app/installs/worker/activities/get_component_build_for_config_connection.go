package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetComponentBuildForConfigConnectionRequest struct {
	ComponentConfigConnectionID string `validate:"required"`
}

// GetComponentBuildForConfigConnection resolves the build for a specific
// config connection. Returns (nil, nil) when none exists yet.
//
// @temporal-gen-v2 activity
// @by-field ComponentConfigConnectionID
func (a *Activities) GetComponentBuildForConfigConnection(ctx context.Context, req GetComponentBuildForConfigConnectionRequest) (*app.ComponentBuild, error) {
	var ccc app.ComponentConfigConnection
	res := a.db.WithContext(ctx).
		Select("id", "component_id", "checksum", "latest_build_id").
		Where("id = ?", req.ComponentConfigConnectionID).
		First(&ccc)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component config connection: %w", res.Error)
	}

	if ccc.LatestBuildID.Valid {
		var pinned app.ComponentBuild
		res = a.db.WithContext(ctx).
			Select("id", "status", "component_config_connection_id").
			Where("id = ?", ccc.LatestBuildID.String).
			First(&pinned)
		if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unable to load pinned component build: %w", res.Error)
		}
		if res.Error == nil && pinned.Status == app.ComponentBuildStatusActive {
			return &pinned, nil
		}
	}

	if ccc.Checksum == "" {
		return nil, nil
	}

	var build app.ComponentBuild
	viewOrTable := views.TableOrViewName(a.db, &app.ComponentConfigConnection{}, "")
	res = a.db.WithContext(ctx).
		Select("component_builds.id", "component_builds.status", "component_builds.component_config_connection_id").
		Joins(fmt.Sprintf("JOIN %s ON %s.id = component_builds.component_config_connection_id", viewOrTable, viewOrTable)).
		Where(fmt.Sprintf("%s.component_id = ?", viewOrTable), ccc.ComponentID).
		Where(fmt.Sprintf("%s.checksum = ?", viewOrTable), ccc.Checksum).
		Where("component_builds.status = ?", app.ComponentBuildStatusActive).
		Order("component_builds.created_at DESC").
		First(&build)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to load active build for config checksum: %w", res.Error)
	}

	return &build, nil
}
