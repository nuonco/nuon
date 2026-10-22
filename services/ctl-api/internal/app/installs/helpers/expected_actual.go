package helpers

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/driftresolution"
)

func (h *Helpers) RecordInstallDeployApplied(ctx context.Context, deployID string, appliedAt time.Time) error {
	var deploy app.InstallDeploy
	if err := h.db.WithContext(ctx).
		Preload("InstallComponent").
		Preload("ComponentBuild.ComponentConfigConnection", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Where(app.InstallDeploy{ID: deployID}).
		First(&deploy).Error; err != nil {
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := driftresolution.LockInstall(ctx, tx, deploy.InstallComponent.InstallID); err != nil {
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
		switch deploy.Type {
		case app.InstallDeployTypeRecover:
			return nil
		case app.InstallDeployTypeTeardown:
			if err := tx.WithContext(ctx).
				Model(&app.InstallComponent{}).
				Where(app.InstallComponent{ID: deploy.InstallComponentID}).
				Update("app_config_ref", app.AppConfigRef{
					ExpectedConfigID: deploy.InstallComponent.AppConfigRef.ExpectedConfigID,
				}).Error; err != nil {
				return fmt.Errorf("unable to clear actual deploy on install component: %w", err)
			}
		default:
			appConfigID := deploy.ComponentBuild.ComponentConfigConnection.AppConfigID
			if appConfigID == "" {
				return fmt.Errorf("component build %s has no app config", deploy.ComponentBuildID)
			}

			at := appliedAt
			if err := tx.WithContext(ctx).
				Model(&app.InstallComponent{}).
				Where(app.InstallComponent{ID: deploy.InstallComponentID}).
				Update("app_config_ref", app.AppConfigRef{
					ExpectedConfigID:    appConfigID,
					AppliedConfigID:     appConfigID,
					AppliedConfigAt:     &at,
					AppliedConfigByType: app.AppConfigRefByTypeInstallDeploys,
					AppliedConfigByID:   deploy.ID,
				}).Error; err != nil {
				return fmt.Errorf("unable to record actual deploy on install component: %w", err)
			}
		}
		return nil
	})
}

func (h *Helpers) RecordInstallSandboxRunApplied(ctx context.Context, runID string, appliedAt time.Time) error {
	var run app.InstallSandboxRun
	if err := h.db.WithContext(ctx).
		Where(app.InstallSandboxRun{ID: runID}).
		First(&run).Error; err != nil {
		return fmt.Errorf("unable to get install sandbox run: %w", err)
	}

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (h *Helpers) RecordInstallActionWorkflowRunApplied(ctx context.Context, runID string, appliedAt time.Time) error {
	var run app.InstallActionWorkflowRun
	if err := h.db.WithContext(ctx).
		Preload("ActionWorkflowConfig", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Where(app.InstallActionWorkflowRun{ID: runID}).
		First(&run).Error; err != nil {
		return fmt.Errorf("unable to get install action workflow run: %w", err)
	}
	if run.InstallActionWorkflowID.Empty() {
		return nil
	}
	if run.ActionWorkflowConfigID.Empty() {
		return nil
	}
	if run.ActionWorkflowConfig.AppConfigID == "" {
		return fmt.Errorf("action workflow config %s for run %s has no app config", run.ActionWorkflowConfigID.String, run.ID)
	}

	at := appliedAt
	if err := h.db.WithContext(ctx).
		Model(&app.InstallActionWorkflow{}).
		Where(app.InstallActionWorkflow{ID: run.InstallActionWorkflowID.String}).
		Update("app_config_ref", app.AppConfigRef{
			ExpectedConfigID:    run.ActionWorkflowConfig.AppConfigID,
			AppliedConfigID:     run.ActionWorkflowConfig.AppConfigID,
			AppliedConfigAt:     &at,
			AppliedConfigByType: app.AppConfigRefByTypeInstallActionWorkflowRuns,
			AppliedConfigByID:   run.ID,
		}).Error; err != nil {
		return fmt.Errorf("unable to record applied action workflow run: %w", err)
	}
	return nil
}

func (h *Helpers) RecordInstallStackVersionApplied(ctx context.Context, stackVersion *app.InstallStackVersion, requestType string) error {
	if requestType != PhoneHomeRequestTypeCreate && requestType != PhoneHomeRequestTypeUpdate {
		return nil
	}
	if stackVersion.InstallStackID == "" {
		return nil
	}

	appConfigID := stackVersion.AppConfigID

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := driftresolution.LockInstall(ctx, tx, stackVersion.InstallID); err != nil {
			return err
		}
		var existing app.InstallStack
		if err := tx.WithContext(ctx).
			Select("id", "app_config_ref").
			Where(app.InstallStack{ID: stackVersion.InstallStackID}).
			First(&existing).Error; err != nil {
			return fmt.Errorf("unable to get install stack: %w", err)
		}

		expectedConfigID := existing.AppConfigRef.ExpectedConfigID
		if appConfigID == "" {
			appConfigID = expectedConfigID
		}

		now := time.Now().UTC()
		if existing.AppConfigRef.AppliedConfigByID == stackVersion.ID && existing.AppConfigRef.AppliedConfigAt != nil {
			now = *existing.AppConfigRef.AppliedConfigAt
		}
		ref := app.AppConfigRef{
			ExpectedConfigID:    expectedConfigID,
			AppliedConfigID:     appConfigID,
			AppliedConfigAt:     &now,
			AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
			AppliedConfigByID:   stackVersion.ID,
		}

		if err := tx.WithContext(ctx).
			Model(&app.InstallStack{}).
			Where(app.InstallStack{ID: stackVersion.InstallStackID}).
			Update("app_config_ref", ref).Error; err != nil {
			return fmt.Errorf("unable to record actual stack version on install stack: %w", err)
		}
		return nil
	})
}
