package buildselection

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

func Component(ctx context.Context, db *gorm.DB, cccID string) (*app.ComponentBuild, error) {
	if cccID == "" {
		return nil, fmt.Errorf("component config connection ID is required")
	}
	var ccc app.ComponentConfigConnection
	if err := db.WithContext(ctx).
		Preload("AppConfig").
		Where(app.ComponentConfigConnection{ID: cccID}).
		First(&ccc).Error; err != nil {
		return nil, fmt.Errorf("unable to load component config connection: %w", err)
	}

	if ccc.LatestBuildID.Valid {
		var pinned app.ComponentBuild
		err := db.WithContext(ctx).
			Preload("ComponentConfigConnection").
			Where(app.ComponentBuild{ID: ccc.LatestBuildID.String}).
			First(&pinned).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unable to load pinned component build: %w", err)
		}
		if err == nil && pinned.Status == app.ComponentBuildStatusActive && pinned.ComponentConfigConnection.ComponentID == ccc.ComponentID {
			return &pinned, nil
		}
	}

	if ccc.AppConfig.BuildsPinned {
		return nil, fmt.Errorf("app config %s has no successful selected build for component %s", ccc.AppConfigID, ccc.ComponentID)
	}

	var build app.ComponentBuild
	query := db.WithContext(ctx).
		Select("component_builds.id", "component_builds.status", "component_builds.component_config_connection_id").
		Where(app.ComponentBuild{Status: app.ComponentBuildStatusActive})
	if ccc.Checksum == "" {
		query = query.Where(app.ComponentBuild{ComponentConfigConnectionID: ccc.ID})
	} else {
		viewOrTable := views.TableOrViewName(db, &app.ComponentConfigConnection{}, "")
		query = query.Joins(fmt.Sprintf("JOIN %s ON %s.id = component_builds.component_config_connection_id", viewOrTable, viewOrTable)).
			Where(fmt.Sprintf("%s.component_id = ?", viewOrTable), ccc.ComponentID).
			Where(fmt.Sprintf("%s.checksum = ?", viewOrTable), ccc.Checksum)
	}
	err := query.Order("component_builds.created_at DESC").First(&build).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to load active build for component config: %w", err)
	}
	return &build, nil
}

func ValidateComponent(ctx context.Context, db *gorm.DB, appConfigID, componentID, buildID string) error {
	if appConfigID == "" {
		return nil
	}
	var cfg app.AppConfig
	if err := db.WithContext(ctx).Where(app.AppConfig{ID: appConfigID}).First(&cfg).Error; err != nil {
		return err
	}
	if !cfg.BuildsPinned {
		return nil
	}
	if !cfg.BuildsLocked {
		return fmt.Errorf("app config %s has not completed its builds", appConfigID)
	}
	if componentID == "" || buildID == "" {
		return fmt.Errorf("component ID and build ID are required")
	}
	var ccc app.ComponentConfigConnection
	if err := db.WithContext(ctx).
		Where(app.ComponentConfigConnection{AppConfigID: appConfigID, ComponentID: componentID}).
		First(&ccc).Error; err != nil {
		return err
	}
	selected, err := Component(ctx, db, ccc.ID)
	if err != nil {
		return err
	}
	if selected == nil || selected.ID != buildID {
		return fmt.Errorf("build %s is not selected by app config %s for component %s", buildID, appConfigID, componentID)
	}
	return nil
}

func ValidateDeploy(ctx context.Context, db *gorm.DB, deploy *app.InstallDeploy) error {
	if deploy.Type == app.InstallDeployTypeTeardown {
		return nil
	}
	appConfigID, err := InstallConfigID(ctx, db, &deploy.InstallComponent.Install, generics.FromPtrStr(deploy.InstallWorkflowID))
	if err != nil {
		return err
	}
	return ValidateComponent(ctx, db, appConfigID, deploy.InstallComponent.ComponentID, deploy.ComponentBuildID)
}
