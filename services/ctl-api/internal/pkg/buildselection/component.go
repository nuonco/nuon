package buildselection

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	if build != nil {
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
	var rows []struct {
		ComponentConfigConnectionID string
		SelectedBuildID             string
	}
	if err := db.WithContext(ctx).
		Table("component_config_connections_selected_build_view_v1").
		Select("component_config_connection_id", "selected_build_id").
		Where("component_config_connection_id IN ? AND selected_build_id IS NOT NULL", configIDs).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to load selected component builds: %w", err)
	}
	for _, row := range rows {
		selected[row.ComponentConfigConnectionID] = &app.ComponentBuild{ID: row.SelectedBuildID, Status: app.ComponentBuildStatusActive}
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
	var build app.ComponentBuild
	if err := db.WithContext(ctx).
		Preload("AppBranchRun.Preview").
		Preload("ComponentConfigConnection").
		Where(app.ComponentBuild{ID: buildID}).
		First(&build).Error; err != nil {
		return err
	}
	if build.ComponentConfigConnection.ComponentID == componentID &&
		OnPinnedBranch(&build, cfg.AppBranchID.String) {
		return nil
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

func OnPinnedBranch(build *app.ComponentBuild, branchID string) bool {
	if branchID == "" || build == nil || build.AppBranchRun == nil || build.AppBranchRun.IsPreview() {
		return false
	}
	return build.AppBranchRun.AppBranchID == branchID
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
