package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
)

type ComponentTarget struct {
	Config *app.ComponentConfigConnection
	Build  *app.ComponentBuild
}

func (h *Helpers) EffectiveComponentTargets(ctx context.Context, db *gorm.DB, appConfigID string) (map[string]*ComponentTarget, error) {
	targets := map[string]*ComponentTarget{}
	if appConfigID == "" {
		return targets, nil
	}
	if db == nil {
		db = h.db
	}

	var appCfg app.AppConfig
	if err := db.WithContext(ctx).
		Select("id", "component_ids").
		Where(app.AppConfig{ID: appConfigID}).
		First(&appCfg).Error; err != nil {
		return nil, fmt.Errorf("unable to get app config %s: %w", appConfigID, err)
	}

	var conns []app.ComponentConfigConnection
	if err := db.WithContext(ctx).
		Preload("Component").
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
		targets[conns[i].ComponentID] = &ComponentTarget{Config: &conns[i]}
	}

	buildIDs := make([]string, 0, len(targets))
	for _, target := range targets {
		build, err := buildselection.Component(ctx, db, target.Config.ID)
		if err != nil {
			return nil, err
		}
		target.Build = build
		if build != nil {
			buildIDs = append(buildIDs, build.ID)
		}
	}
	if len(buildIDs) > 0 {
		var builds []app.ComponentBuild
		if err := db.WithContext(ctx).
			Preload("ComponentConfigConnection.Component").
			Where("id IN ?", buildIDs).Find(&builds).Error; err != nil {
			return nil, fmt.Errorf("unable to get selected build configs: %w", err)
		}
		byID := make(map[string]*app.ComponentBuild, len(builds))
		for i := range builds {
			byID[builds[i].ID] = &builds[i]
		}
		for _, target := range targets {
			if target.Build != nil {
				target.Build = byID[target.Build.ID]
			}
		}
	}
	return targets, nil
}

func (h *Helpers) AppliedInstallDeploys(ctx context.Context, db *gorm.DB, installComponents []app.InstallComponent) (map[string]*app.InstallDeploy, error) {
	applied := map[string]*app.InstallDeploy{}
	if len(installComponents) == 0 {
		return applied, nil
	}
	if db == nil {
		db = h.db
	}

	refDeployIDs := make([]string, 0, len(installComponents))
	legacyIDs := make([]string, 0)
	for i := range installComponents {
		ref := installComponents[i].AppConfigRef
		if ref.AppliedConfigByID != "" && ref.AppliedConfigByType == app.AppConfigRefByTypeInstallDeploys {
			if ref.AppliedConfigID != "" {
				refDeployIDs = append(refDeployIDs, ref.AppliedConfigByID)
			}
			continue
		}
		if ref.AppliedConfigByID == "" {
			legacyIDs = append(legacyIDs, installComponents[i].ID)
		}
	}

	if len(legacyIDs) > 0 {
		var latest []string
		if err := db.WithContext(ctx).Raw(`
			SELECT DISTINCT ON (install_component_id) id
			FROM install_deploys
			WHERE deleted_at = 0
			  AND install_component_id IN ?
			  AND type IS DISTINCT FROM ?
			  AND NOT EXISTS (
				SELECT 1 FROM install_workflows w
				WHERE w.id = install_deploys.install_workflow_id AND w.plan_only = true
			  )
			  AND (
				applied_at IS NOT NULL
				OR (type = ? AND COALESCE(NULLIF(status_v2->>'status', ''), status) = ?)
			  )
			ORDER BY install_component_id, COALESCE(applied_at, updated_at) DESC, created_at DESC`,
			legacyIDs, app.InstallDeployTypeRecover, app.InstallDeployTypeSync, app.InstallDeployStatusActive,
		).Scan(&latest).Error; err != nil {
			return nil, fmt.Errorf("unable to get latest applied install deploys: %w", err)
		}
		refDeployIDs = append(refDeployIDs, latest...)
	}
	if len(refDeployIDs) == 0 {
		return applied, nil
	}

	var deploys []app.InstallDeploy
	if err := db.WithContext(ctx).
		Preload("ComponentBuild").
		Preload("ComponentBuild.ComponentConfigConnection", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Where("id IN ?", refDeployIDs).
		Find(&deploys).Error; err != nil {
		return nil, fmt.Errorf("unable to get applied install deploys: %w", err)
	}
	for i := range deploys {
		if deploys[i].Type == app.InstallDeployTypeTeardown {
			continue
		}
		applied[deploys[i].InstallComponentID] = &deploys[i]
	}
	return applied, nil
}

func (h *Helpers) AppliedInstallSandboxRun(ctx context.Context, db *gorm.DB, sandbox *app.InstallSandbox) (*app.InstallSandboxRun, error) {
	if sandbox == nil || sandbox.ID == "" {
		return nil, nil
	}
	if db == nil {
		db = h.db
	}
	ref := sandbox.AppConfigRef
	query := db.WithContext(ctx).
		Preload("AppSandboxConfig", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		})
	switch {
	case ref.AppliedConfigByType == app.AppConfigRefByTypeInstallSandboxRuns && ref.AppliedConfigByID != "":
		if ref.AppliedConfigID == "" {
			return nil, nil
		}
		query = query.Where(app.InstallSandboxRun{ID: ref.AppliedConfigByID})
	case ref.AppliedConfigByID == "":
		query = query.
			Where(app.InstallSandboxRun{InstallSandboxID: &sandbox.ID}).
			Where("applied_at IS NOT NULL").
			Order("applied_at DESC").
			Order("created_at DESC")
	default:
		return nil, nil
	}

	var run app.InstallSandboxRun
	err := query.First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get applied install sandbox run: %w", err)
	}
	if run.RunType == app.SandboxRunTypeDeprovision {
		return nil, nil
	}
	return &run, nil
}
