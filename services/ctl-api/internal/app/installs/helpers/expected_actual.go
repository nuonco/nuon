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
		Update("expected_app_config_id", appConfigID).Error; err != nil {
		return fmt.Errorf("unable to set expected app config on install stack: %w", err)
	}

	var sandboxConfigIDs []string
	if err := h.db.WithContext(ctx).
		Model(&app.AppSandboxConfig{}).
		Where(app.AppSandboxConfig{AppConfigID: appConfigID}).
		Pluck("id", &sandboxConfigIDs).Error; err != nil {
		return fmt.Errorf("unable to get app sandbox config: %w", err)
	}
	var expectedSandboxConfigID *string
	if len(sandboxConfigIDs) == 1 {
		expectedSandboxConfigID = &sandboxConfigIDs[0]
	}
	if err := h.db.WithContext(ctx).
		Model(&app.InstallSandbox{}).
		Where(app.InstallSandbox{InstallID: installID}).
		Update("expected_app_sandbox_config_id", expectedSandboxConfigID).Error; err != nil {
		return fmt.Errorf("unable to set expected sandbox config on install sandbox: %w", err)
	}

	return nil
}

func (h *Helpers) RecordInstallDeployApplied(ctx context.Context, deployID string, appliedAt time.Time) error {
	var deploy app.InstallDeploy
	if err := h.db.WithContext(ctx).
		Where(app.InstallDeploy{ID: deployID}).
		First(&deploy).Error; err != nil {
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	var updates map[string]any
	switch deploy.Type {
	case app.InstallDeployTypeRecover:
		return nil
	case app.InstallDeployTypeTeardown:
		updates = map[string]any{
			"actual_install_deploy_id":  nil,
			"actual_component_build_id": nil,
			"actual_applied_at":         nil,
		}
	default:
		updates = map[string]any{
			"actual_install_deploy_id":  deploy.ID,
			"actual_component_build_id": deploy.ComponentBuildID,
			"actual_applied_at":         appliedAt,
		}
	}

	if err := h.db.WithContext(ctx).
		Model(&app.InstallComponent{}).
		Where(app.InstallComponent{ID: deploy.InstallComponentID}).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("unable to record actual deploy on install component: %w", err)
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

	var updates map[string]any
	switch run.RunType {
	case app.SandboxRunTypeProvision, app.SandboxRunTypeReprovision:
		updates = map[string]any{
			"actual_app_sandbox_config_id":  run.AppSandboxConfigID,
			"actual_install_sandbox_run_id": run.ID,
			"actual_applied_at":             appliedAt,
		}
	case app.SandboxRunTypeDeprovision:
		updates = map[string]any{
			"actual_app_sandbox_config_id":  nil,
			"actual_install_sandbox_run_id": nil,
			"actual_applied_at":             nil,
		}
	default:
		return nil
	}

	if err := h.db.WithContext(ctx).
		Model(&app.InstallSandbox{}).
		Where(app.InstallSandbox{ID: *run.InstallSandboxID}).
		Updates(updates).Error; err != nil {
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

	updates := map[string]any{
		"actual_install_stack_version_id": stackVersion.ID,
		"actual_applied_at":               time.Now().UTC(),
		"actual_app_config_id":            nil,
	}
	if stackVersion.AppConfigID != "" {
		updates["actual_app_config_id"] = stackVersion.AppConfigID
	}

	if err := h.db.WithContext(ctx).
		Model(&app.InstallStack{}).
		Where(app.InstallStack{ID: stackVersion.InstallStackID}).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("unable to record actual stack version on install stack: %w", err)
	}
	return nil
}
