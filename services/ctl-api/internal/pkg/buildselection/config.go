package buildselection

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func InstallConfigID(ctx context.Context, db *gorm.DB, install *app.Install, workflowID string) (string, error) {
	appConfigID := install.AppConfigID
	if workflowID == "" {
		return appConfigID, nil
	}
	var flow app.Workflow
	if err := db.WithContext(ctx).Where(app.Workflow{ID: workflowID}).First(&flow).Error; err != nil {
		return "", err
	}
	if flow.Type == app.WorkflowTypeAppBranchConfigUpdate || flow.Type == app.WorkflowTypeReconcileAppVersion {
		if target := generics.FromPtrStr(flow.Metadata["new_app_config_id"]); target != "" {
			appConfigID = target
		}
	}
	return appConfigID, nil
}

func WithMutableConfig(ctx context.Context, db *gorm.DB, appConfigID string, fn func(*gorm.DB, *app.AppConfig) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cfg app.AppConfig
		if err := tx.Scopes(scopes.WithDisableViews).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.AppConfig{ID: appConfigID}).First(&cfg).Error; err != nil {
			return err
		}
		if cfg.BuildsLocked {
			return stderr.ErrUser{
				Err:         fmt.Errorf("build selections for app config %s are locked", appConfigID),
				Description: "Trigger a new app branch run to rebuild components and sandbox.",
			}
		}
		return fn(tx, &cfg)
	})
}

func PinSandbox(ctx context.Context, db *gorm.DB, appConfigID, buildID string) error {
	return WithMutableConfig(ctx, db, appConfigID, func(tx *gorm.DB, cfg *app.AppConfig) error {
		if !cfg.BuildsPinned {
			return nil
		}
		var build app.AppSandboxBuild
		if err := tx.Where(app.AppSandboxBuild{ID: buildID, AppID: cfg.AppID}).First(&build).Error; err != nil {
			return err
		}
		return tx.Model(&app.AppConfig{}).Where(app.AppConfig{ID: cfg.ID}).
			Update("sandbox_build_id", generics.NewNullString(buildID)).Error
	})
}

func FreezeInTx(ctx context.Context, tx *gorm.DB, appConfigID string) error {
	var cfg app.AppConfig
	if err := tx.WithContext(ctx).Scopes(scopes.WithDisableViews).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(app.AppConfig{ID: appConfigID}).First(&cfg).Error; err != nil {
		return err
	}
	if !cfg.BuildsPinned || cfg.BuildsLocked {
		return nil
	}
	var conns []app.ComponentConfigConnection
	if err := tx.WithContext(ctx).Where(app.ComponentConfigConnection{AppConfigID: appConfigID}).Find(&conns).Error; err != nil {
		return err
	}
	for _, componentID := range cfg.ComponentIDs {
		found := false
		for _, ccc := range conns {
			if ccc.ComponentID != componentID {
				continue
			}
			if _, err := Component(ctx, tx, ccc.ID); err != nil {
				return err
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("app config %s has no config connection for component %s", appConfigID, componentID)
		}
	}
	var sandbox app.AppSandboxConfig
	if err := tx.WithContext(ctx).Where(app.AppSandboxConfig{AppConfigID: appConfigID}).First(&sandbox).Error; err != nil {
		return err
	}
	if _, err := Sandbox(ctx, tx, &cfg); err != nil {
		return err
	}
	return tx.WithContext(ctx).Model(&app.AppConfig{}).Where(app.AppConfig{ID: appConfigID}).Update("builds_locked", true).Error
}

func Sandbox(ctx context.Context, db *gorm.DB, cfg *app.AppConfig) (*app.AppSandboxBuild, error) {
	if !cfg.SandboxBuildID.Valid || cfg.SandboxBuildID.String == "" {
		return nil, fmt.Errorf("app config %s has no selected sandbox build", cfg.ID)
	}
	var build app.AppSandboxBuild
	if err := db.WithContext(ctx).
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("VCSConnectionCommit").
		Where(app.AppSandboxBuild{ID: cfg.SandboxBuildID.String, AppID: cfg.AppID}).
		First(&build).Error; err != nil {
		return nil, err
	}
	if build.Status != app.AppSandboxBuildStatusActive {
		return nil, fmt.Errorf("selected sandbox build %s for app config %s is not successful", build.ID, cfg.ID)
	}
	return &build, nil
}
