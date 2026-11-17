package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type ComponentTarget struct {
	Config *app.ComponentConfigConnection
	Build  *app.ComponentBuild
}

var componentConfigFields = []string{"id", "app_config_id", "component_id", "latest_build_id", "checksum", "references", "component_dependency_ids", "toggleable", "default_enabled"}

func componentConfigLoad(db *gorm.DB) *gorm.DB {
	return db.Scopes(scopes.WithDisableViews).Select(componentConfigFields).
		Where("EXISTS (SELECT 1 FROM app_configs a WHERE a.id = component_config_connections.app_config_id)")
}

func componentComponentLoad(db *gorm.DB) *gorm.DB { return db.Select("id", "name", "type") }

func appliedDeployConfigLoad(db *gorm.DB) *gorm.DB { return db.Unscoped() }

func (h *Helpers) EffectiveComponentTargets(ctx context.Context, appConfigID string) (map[string]*ComponentTarget, error) {
	targets, err := h.EffectiveComponentTargetConfigs(ctx, appConfigID)
	if err != nil {
		return nil, err
	}

	configs := make([]app.ComponentConfigConnection, 0, len(targets))
	for _, target := range targets {
		configs = append(configs, *target.Config)
	}

	selected, err := buildselection.Components(ctx, h.db, configs)
	if err != nil {
		return nil, err
	}
	buildIDs := make([]string, 0, len(targets))
	for _, target := range targets {
		build := selected[target.Config.ID]
		target.Build = build
		if build != nil {
			buildIDs = append(buildIDs, build.ID)
		}
	}
	builds, err := h.ComponentBuildsWithConfigs(ctx, buildIDs)
	if err != nil {
		return nil, err
	}
	for _, target := range targets {
		if target.Build != nil {
			target.Build = builds[target.Build.ID]
		}
	}
	return targets, nil
}

func (h *Helpers) EffectiveComponentTargetConfigs(ctx context.Context, appConfigID string) (map[string]*ComponentTarget, error) {
	targets := map[string]*ComponentTarget{}
	if appConfigID == "" {
		return targets, nil
	}

	var appCfg app.AppConfig
	if err := h.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Select("id", "component_ids", "builds_pinned").
		Where(app.AppConfig{ID: appConfigID}).
		First(&appCfg).Error; err != nil {
		return nil, fmt.Errorf("unable to get app config %s: %w", appConfigID, err)
	}

	var conns []app.ComponentConfigConnection
	if err := componentConfigLoad(h.db.WithContext(ctx)).
		Preload("Component", componentComponentLoad).
		Where(app.ComponentConfigConnection{AppConfigID: appConfigID}).
		Find(&conns).Error; err != nil {
		return nil, fmt.Errorf("unable to get component configs for app config %s: %w", appConfigID, err)
	}

	member := make(map[string]struct{}, len(appCfg.ComponentIDs))
	for _, id := range appCfg.ComponentIDs {
		member[id] = struct{}{}
	}
	for i := range conns {
		if len(member) > 0 {
			if _, ok := member[conns[i].ComponentID]; !ok {
				continue
			}
		}
		conns[i].AppConfig = appCfg
		targets[conns[i].ComponentID] = &ComponentTarget{Config: &conns[i]}
	}
	return targets, nil
}

func (h *Helpers) ComponentBuildsWithConfigs(ctx context.Context, buildIDs []string) (map[string]*app.ComponentBuild, error) {
	builds := map[string]*app.ComponentBuild{}
	if len(buildIDs) == 0 {
		return builds, nil
	}
	var rows []app.ComponentBuild
	if err := h.db.WithContext(ctx).
		Select("id", "status", "status_v2", "component_config_connection_id", "source_checksum", "source_digest", "source_identity").
		Preload("ComponentConfigConnection", componentConfigLoad).
		Preload("ComponentConfigConnection.Component", componentComponentLoad).
		Where("id IN ?", buildIDs).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to get selected build configs: %w", err)
	}
	for i := range rows {
		builds[rows[i].ID] = &rows[i]
	}
	return builds, nil
}

func (h *Helpers) AppliedInstallDeploys(ctx context.Context, installComponents []app.InstallComponent) (map[string]*app.InstallDeploy, error) {
	applied := map[string]*app.InstallDeploy{}
	if len(installComponents) == 0 {
		return applied, nil
	}

	ids := make([]string, 0, len(installComponents))
	installIDs := map[string]struct{}{}
	for i := range installComponents {
		ids = append(ids, installComponents[i].ID)
		installIDs[installComponents[i].InstallID] = struct{}{}
	}

	var deploys []*app.InstallDeploy
	if err := h.db.WithContext(ctx).
		Scopes(scopes.WithOverrideTable("install_deploys_live_view_v1")).
		Preload("ComponentBuild").
		Preload("ComponentBuild.ComponentConfigConnection", appliedDeployConfigLoad).
		Where("install_id IN ? AND install_component_id IN ?", generics.MapToKeys(installIDs), ids).
		Find(&deploys).Error; err != nil {
		return nil, fmt.Errorf("unable to get applied install deploys: %w", err)
	}
	for i := range deploys {
		applied[deploys[i].InstallComponentID] = deploys[i]
	}
	return applied, nil
}

func (h *Helpers) AppliedInstallDeploysByIDs(ctx context.Context, deployIDs []string) (map[string]*app.InstallDeploy, error) {
	applied := map[string]*app.InstallDeploy{}
	if len(deployIDs) == 0 {
		return applied, nil
	}

	var deploys []*app.InstallDeploy
	if err := h.db.WithContext(ctx).
		Preload("ComponentBuild").
		Preload("ComponentBuild.ComponentConfigConnection", appliedDeployConfigLoad).
		Where("id IN ?", deployIDs).
		Find(&deploys).Error; err != nil {
		return nil, fmt.Errorf("unable to get applied install deploys: %w", err)
	}
	for i := range deploys {
		applied[deploys[i].InstallComponentID] = deploys[i]
	}
	return applied, nil
}

func (h *Helpers) AppliedInstallSandboxRun(ctx context.Context, sandbox *app.InstallSandbox) (*app.InstallSandboxRun, error) {
	if sandbox == nil || sandbox.ID == "" {
		return nil, nil
	}
	var run app.InstallSandboxRun
	err := h.db.WithContext(ctx).
		Scopes(scopes.WithOverrideTable("install_sandbox_runs_live_view_v1")).
		Preload("AppSandboxConfig", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Preload("InstallWorkflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "type", "metadata")
		}).
		Where(app.InstallSandboxRun{InstallSandboxID: &sandbox.ID}).
		First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get applied install sandbox run: %w", err)
	}
	return &run, nil
}

func (h *Helpers) AppliedBuildIDs(ctx context.Context, installID string) (map[string]string, error) {
	var components []app.InstallComponent
	if err := h.db.WithContext(ctx).Where(app.InstallComponent{InstallID: installID}).Find(&components).Error; err != nil {
		return nil, fmt.Errorf("unable to get install components: %w", err)
	}
	deploys, err := h.AppliedInstallDeploys(ctx, components)
	if err != nil {
		return nil, err
	}
	ids := map[string]string{"sandbox": ""}
	for i := range components {
		if deploy := deploys[components[i].ID]; deploy != nil {
			ids[components[i].ComponentID] = deploy.ComponentBuildID
		}
	}

	var sandbox app.InstallSandbox
	err = h.db.WithContext(ctx).Where(app.InstallSandbox{InstallID: installID}).First(&sandbox).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ids, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get install sandbox: %w", err)
	}
	run, err := h.AppliedInstallSandboxRun(ctx, &sandbox)
	if err != nil {
		return nil, err
	}
	if run != nil {
		if ids["sandbox"], err = h.AppliedSandboxBuildID(ctx, run); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func (h *Helpers) AppliedSandboxBuildID(ctx context.Context, run *app.InstallSandboxRun) (string, error) {
	if run.AppSandboxBuildID != "" {
		return run.AppSandboxBuildID, nil
	}
	var cfg app.AppConfig
	if err := h.db.WithContext(ctx).Unscoped().Where(app.AppConfig{ID: run.AppSandboxConfig.AppConfigID}).First(&cfg).Error; err != nil {
		return "", err
	}
	if cfg.BuildsPinned && cfg.BuildsLocked && cfg.SandboxBuildID.Valid {
		return cfg.SandboxBuildID.String, nil
	}
	return "", nil
}
