package appliedstate

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/driftresolution"
)

func RecordDeploy(ctx context.Context, db *gorm.DB, deployID string) error {
	appliedAt := time.Now().UTC()
	var deploy app.InstallDeploy
	if err := db.WithContext(ctx).
		Preload("InstallComponent", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Preload("ComponentBuild.ComponentConfigConnection", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Where(app.InstallDeploy{ID: deployID}).
		First(&deploy).Error; err != nil {
		return fmt.Errorf("unable to get install deploy: %w", err)
	}
	installComponent := deploy.InstallComponent
	if installComponent == nil {
		return fmt.Errorf("install component %s not found for deploy %s", deploy.InstallComponentID, deploy.ID)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := driftresolution.LockInstall(ctx, tx, installComponent.InstallID); err != nil {
			return err
		}
		if err := tx.Select("applied_at").Where(app.InstallDeploy{ID: deploy.ID}).First(&deploy).Error; err != nil {
			return err
		}
		if deploy.AppliedAt != nil {
			appliedAt = *deploy.AppliedAt
		}
		if err := tx.Model(&app.InstallDeploy{}).Where(app.InstallDeploy{ID: deploy.ID}).Update("applied_at", appliedAt).Error; err != nil {
			return err
		}
		var ref app.AppConfigRef
		operation := "record"
		switch deploy.Type {
		case app.InstallDeployTypeRecover:
			return nil
		case app.InstallDeployTypeTeardown:
			operation = "clear"
			ref = app.AppConfigRef{
				ExpectedConfigID: installComponent.AppConfigRef.ExpectedConfigID,
			}
		default:
			appConfigID := deploy.AppConfigID
			if appConfigID == "" && deploy.InstallWorkflowID != nil {
				var flow app.Workflow
				if err := tx.WithContext(ctx).Where(app.Workflow{ID: *deploy.InstallWorkflowID}).First(&flow).Error; err != nil {
					return fmt.Errorf("unable to get deploy workflow: %w", err)
				}
				if flow.Type == app.WorkflowTypeAppBranchConfigUpdate && flow.Metadata["new_app_config_id"] != nil {
					appConfigID = *flow.Metadata["new_app_config_id"]
				}
			}
			if appConfigID == "" && deploy.ComponentBuild != nil {
				appConfigID = deploy.ComponentBuild.ComponentConfigConnection.AppConfigID
			}
			if appConfigID == "" {
				return fmt.Errorf("component build %s has no app config", deploy.ComponentBuildID)
			}

			at := appliedAt
			ref = app.AppConfigRef{
				ExpectedConfigID:    appConfigID,
				AppliedConfigID:     appConfigID,
				AppliedConfigAt:     &at,
				AppliedConfigByType: app.AppConfigRefByTypeInstallDeploys,
				AppliedConfigByID:   deploy.ID,
			}
		}
		if err := tx.WithContext(ctx).
			Model(&app.InstallComponent{}).
			Where(app.InstallComponent{ID: deploy.InstallComponentID}).
			Update("app_config_ref", ref).Error; err != nil {
			return fmt.Errorf("unable to %s actual deploy on install component: %w", operation, err)
		}
		return nil
	})
}

func RecordSandboxRun(ctx context.Context, db *gorm.DB, runID string) error {
	appliedAt := time.Now().UTC()
	var run app.InstallSandboxRun
	if err := db.WithContext(ctx).
		Where(app.InstallSandboxRun{ID: runID}).
		First(&run).Error; err != nil {
		return fmt.Errorf("unable to get install sandbox run: %w", err)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := driftresolution.LockInstall(ctx, tx, run.InstallID); err != nil {
			return err
		}
		if err := tx.Select("applied_at").Where(app.InstallSandboxRun{ID: run.ID}).First(&run).Error; err != nil {
			return err
		}
		if run.AppliedAt != nil {
			appliedAt = *run.AppliedAt
		}
		if err := tx.Model(&app.InstallSandboxRun{}).Where(app.InstallSandboxRun{ID: run.ID}).Update("applied_at", appliedAt).Error; err != nil {
			return err
		}
		if run.InstallSandboxID == nil || *run.InstallSandboxID == "" {
			return nil
		}
		var ref app.AppConfigRef
		switch run.RunType {
		case app.SandboxRunTypeProvision, app.SandboxRunTypeReprovision:
			if run.AppSandboxConfigID == "" {
				return nil
			}
			var sandboxConfig app.AppSandboxConfig
			if err := tx.WithContext(ctx).Unscoped().
				Select("id", "app_config_id").
				Where(app.AppSandboxConfig{ID: run.AppSandboxConfigID}).
				First(&sandboxConfig).Error; err != nil {
				return fmt.Errorf("unable to get app sandbox config for sandbox run: %w", err)
			}
			at := appliedAt
			ref = app.AppConfigRef{
				ExpectedConfigID:    sandboxConfig.AppConfigID,
				AppliedConfigID:     sandboxConfig.AppConfigID,
				AppliedConfigAt:     &at,
				AppliedConfigByType: app.AppConfigRefByTypeInstallSandboxRuns,
				AppliedConfigByID:   run.ID,
			}
		case app.SandboxRunTypeDeprovision:
			var existing app.InstallSandbox
			if err := tx.WithContext(ctx).
				Select("id", "app_config_ref").
				Where(app.InstallSandbox{ID: *run.InstallSandboxID}).
				First(&existing).Error; err != nil {
				return fmt.Errorf("unable to get install sandbox: %w", err)
			}
			ref = app.AppConfigRef{
				ExpectedConfigID: existing.AppConfigRef.ExpectedConfigID,
			}
		default:
			return nil
		}

		if err := tx.WithContext(ctx).
			Model(&app.InstallSandbox{}).
			Where(app.InstallSandbox{ID: *run.InstallSandboxID}).
			Update("app_config_ref", ref).Error; err != nil {
			return fmt.Errorf("unable to record actual sandbox run on install sandbox: %w", err)
		}
		return nil
	})
}
