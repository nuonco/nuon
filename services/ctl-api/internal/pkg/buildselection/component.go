package buildselection

import (
	"context"
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

	builds, err := Components(ctx, db, []app.ComponentConfigConnection{ccc})
	if err != nil {
		return nil, err
	}
	build := builds[ccc.ID]
	if build == nil && ccc.AppConfig.BuildsPinned {
		return nil, fmt.Errorf("app config %s has no successful selected build for component %s", ccc.AppConfigID, ccc.ComponentID)
	}
	if build != nil && build.ComponentConfigConnection.ID != "" {
		var pinned app.ComponentBuild
		if err := db.WithContext(ctx).
			Preload("ComponentConfigConnection").
			Where(app.ComponentBuild{ID: build.ID}).
			First(&pinned).Error; err != nil {
			return nil, fmt.Errorf("unable to load pinned component build: %w", err)
		}
		return &pinned, nil
	}
	return build, nil
}

func Components(ctx context.Context, db *gorm.DB, configs []app.ComponentConfigConnection) (map[string]*app.ComponentBuild, error) {
	selected := make(map[string]*app.ComponentBuild, len(configs))
	if len(configs) == 0 {
		return selected, nil
	}
	configIDs := make([]string, 0, len(configs))
	for _, cfg := range configs {
		configIDs = append(configIDs, cfg.ID)
	}
	liveBuild := "b.deleted_at = 0 AND "
	liveConfig := "bc.deleted_at = 0 AND "
	if db.Statement.Unscoped {
		liveBuild = ""
		liveConfig = ""
	}
	configAppJoin := "JOIN app_configs a ON a.id = bc.app_config_id"
	legacyAppJoin := "JOIN app_configs a ON a.id = old.app_config_id"
	if views.TableOrViewName(db, &app.ComponentConfigConnection{}, "") == "component_config_connections" {
		configAppJoin = ""
		legacyAppJoin = ""
	}
	var pins []struct {
		TargetConfigID string
		BuildID        string
		BuildConfigID  string
		ComponentID    string
	}
	if err := db.WithContext(ctx).Raw(`
		SELECT c.id AS target_config_id, b.id AS build_id, b.component_config_connection_id AS build_config_id, bc.component_id
		FROM component_config_connections c
		JOIN component_builds b ON b.id = c.latest_build_id
		JOIN component_config_connections bc ON bc.id = b.component_config_connection_id
		`+configAppJoin+`
		WHERE c.id IN ? AND `+liveBuild+liveConfig+`COALESCE(NULLIF(b.status_v2->>'status', ''), b.status) = ? AND bc.component_id = c.component_id`,
		configIDs, app.ComponentBuildStatusActive).Scan(&pins).Error; err != nil {
		return nil, fmt.Errorf("unable to load pinned component builds: %w", err)
	}
	for _, pin := range pins {
		selected[pin.TargetConfigID] = &app.ComponentBuild{
			ID: pin.BuildID, Status: app.ComponentBuildStatusActive, ComponentConfigConnectionID: pin.BuildConfigID,
			ComponentConfigConnection: app.ComponentConfigConnection{ID: pin.BuildConfigID, ComponentID: pin.ComponentID},
		}
	}
	legacyIDs := make([]string, 0, len(configs))
	for _, cfg := range configs {
		if selected[cfg.ID] != nil {
			continue
		}
		if cfg.AppConfig.BuildsPinned {
			continue
		}
		legacyIDs = append(legacyIDs, cfg.ID)
	}
	if len(legacyIDs) == 0 {
		return selected, nil
	}
	var legacy []struct {
		TargetConfigID string
		BuildID        string
		BuildConfigID  string
	}
	if err := db.WithContext(ctx).Raw(`
		WITH wanted AS MATERIALIZED (
			SELECT id, component_id, checksum FROM component_config_connections WHERE id IN ?
		), candidate_configs AS (
			SELECT id AS target_config_id, id AS build_config_id FROM wanted WHERE COALESCE(checksum, '') = ''
			UNION ALL
			SELECT w.id, old.id FROM wanted w
			JOIN component_config_connections old ON old.component_id = w.component_id AND old.checksum = w.checksum
			`+legacyAppJoin+`
			WHERE COALESCE(w.checksum, '') <> ''
		)
		SELECT DISTINCT ON (c.target_config_id) c.target_config_id, b.id AS build_id, b.component_config_connection_id AS build_config_id
		FROM candidate_configs c JOIN component_builds b ON b.component_config_connection_id = c.build_config_id
		WHERE `+liveBuild+`b.status = ?
		ORDER BY c.target_config_id, b.created_at DESC, b.id ASC`, legacyIDs, app.ComponentBuildStatusActive).
		Scan(&legacy).Error; err != nil {
		return nil, fmt.Errorf("unable to load active build for component config: %w", err)
	}
	for _, build := range legacy {
		selected[build.TargetConfigID] = &app.ComponentBuild{
			ID: build.BuildID, Status: app.ComponentBuildStatusActive, ComponentConfigConnectionID: build.BuildConfigID,
		}
	}
	return selected, nil
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
	installComponent := deploy.InstallComponent
	if installComponent == nil || installComponent.Install == nil {
		return fmt.Errorf("deploy %s has no install component or install loaded", deploy.ID)
	}
	appConfigID, err := InstallConfigID(ctx, db, installComponent.Install, generics.FromPtrStr(deploy.InstallWorkflowID))
	if err != nil {
		return err
	}
	return ValidateComponent(ctx, db, appConfigID, installComponent.ComponentID, deploy.ComponentBuildID)
}
