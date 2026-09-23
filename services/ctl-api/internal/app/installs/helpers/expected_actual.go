package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) SetInstallExpectedAppConfig(ctx context.Context, installID, appConfigID string) error {
	if err := h.db.WithContext(ctx).
		Model(&app.InstallStack{}).
		Where(app.InstallStack{InstallID: installID}).
		Update("app_config_ref", app.AppConfigRef{
			ExpectedConfigID: appConfigID,
		}).Error; err != nil {
		return fmt.Errorf("unable to set expected app config on install stack: %w", err)
	}

	if err := h.db.WithContext(ctx).
		Model(&app.InstallSandbox{}).
		Where(app.InstallSandbox{InstallID: installID}).
		Update("app_config_ref", app.AppConfigRef{
			ExpectedConfigID: appConfigID,
		}).Error; err != nil {
		return fmt.Errorf("unable to set expected config on install sandbox: %w", err)
	}

	return nil
}

func (h *Helpers) RecordInstallDeployApplied(ctx context.Context, deployID string, appliedAt time.Time) error {
	var deploy app.InstallDeploy
	if err := h.db.WithContext(ctx).
		Preload("InstallComponent").
		Where(app.InstallDeploy{ID: deployID}).
		First(&deploy).Error; err != nil {
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	switch deploy.Type {
	case app.InstallDeployTypeRecover:
		return nil
	case app.InstallDeployTypeTeardown:
		if err := h.db.WithContext(ctx).
			Model(&app.InstallComponent{}).
			Where(app.InstallComponent{ID: deploy.InstallComponentID}).
			Update("app_config_ref", app.AppConfigRef{
				ExpectedConfigID: deploy.InstallComponent.AppConfigRef.ExpectedConfigID,
			}).Error; err != nil {
			return fmt.Errorf("unable to clear actual deploy on install component: %w", err)
		}
	default:
		var install app.Install
		if err := h.db.WithContext(ctx).
			Select("id", "app_config_id").
			Where(app.Install{ID: deploy.InstallComponent.InstallID}).
			First(&install).Error; err != nil {
			return fmt.Errorf("unable to get install for deploy: %w", err)
		}

		at := appliedAt
		if err := h.db.WithContext(ctx).
			Model(&app.InstallComponent{}).
			Where(app.InstallComponent{ID: deploy.InstallComponentID}).
			Update("app_config_ref", app.AppConfigRef{
				ExpectedConfigID:    install.AppConfigID,
				AppliedConfigID:     install.AppConfigID,
				AppliedConfigAt:     &at,
				AppliedConfigByType: app.AppConfigRefByTypeInstallDeploys,
				AppliedConfigByID:   deploy.ID,
			}).Error; err != nil {
			return fmt.Errorf("unable to record actual deploy on install component: %w", err)
		}
	}
	return nil
}

func (h *Helpers) RecordInstallSandboxRunApplied(ctx context.Context, runID string, appliedAt time.Time) error {
	var run app.InstallSandboxRun
	if err := h.db.WithContext(ctx).
		Where(app.InstallSandboxRun{ID: runID}).
		First(&run).Error; err != nil {
		return fmt.Errorf("unable to get install sandbox run: %w", err)
	}
	if run.InstallSandboxID == nil || *run.InstallSandboxID == "" {
		return nil
	}

	var ref app.AppConfigRef
	switch run.RunType {
	case app.SandboxRunTypeProvision, app.SandboxRunTypeReprovision:
		var install app.Install
		if err := h.db.WithContext(ctx).
			Select("id", "app_config_id").
			Where(app.Install{ID: run.InstallID}).
			First(&install).Error; err != nil {
			return fmt.Errorf("unable to get install for sandbox run: %w", err)
		}
		at := appliedAt
		ref = app.AppConfigRef{
			ExpectedConfigID:    install.AppConfigID,
			AppliedConfigID:     install.AppConfigID,
			AppliedConfigAt:     &at,
			AppliedConfigByType: app.AppConfigRefByTypeInstallSandboxRuns,
			AppliedConfigByID:   run.ID,
		}
	case app.SandboxRunTypeDeprovision:
		var existing app.InstallSandbox
		if err := h.db.WithContext(ctx).
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

	if err := h.db.WithContext(ctx).
		Model(&app.InstallSandbox{}).
		Where(app.InstallSandbox{ID: *run.InstallSandboxID}).
		Update("app_config_ref", ref).Error; err != nil {
		return fmt.Errorf("unable to record actual sandbox run on install sandbox: %w", err)
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

	var existing app.InstallStack
	if err := h.db.WithContext(ctx).
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
	ref := app.AppConfigRef{
		ExpectedConfigID:    expectedConfigID,
		AppliedConfigID:     appConfigID,
		AppliedConfigAt:     &now,
		AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
		AppliedConfigByID:   stackVersion.ID,
	}

	if err := h.db.WithContext(ctx).
		Model(&app.InstallStack{}).
		Where(app.InstallStack{ID: stackVersion.InstallStackID}).
		Update("app_config_ref", ref).Error; err != nil {
		return fmt.Errorf("unable to record actual stack version on install stack: %w", err)
	}
	return nil
}
