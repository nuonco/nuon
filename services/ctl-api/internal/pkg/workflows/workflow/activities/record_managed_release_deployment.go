package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type RecordManagedReleaseDeploymentRequest struct {
	WorkflowID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) RecordManagedReleaseDeployment(ctx context.Context, req RecordManagedReleaseDeploymentRequest) error {
	var flow app.Workflow
	if err := a.db.WithContext(ctx).Where(app.Workflow{ID: req.WorkflowID}).First(&flow).Error; err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}
	if flow.OwnerType != "installs" || flow.PlanOnly || !recordsReleaseDeployment(flow.Type) {
		return nil
	}

	var existing int64
	if err := a.db.WithContext(ctx).Model(&app.InstallReleaseDeployment{}).Where(app.InstallReleaseDeployment{
		InstallID: flow.OwnerID, OperationID: flow.ID,
	}).Count(&existing).Error; err != nil {
		return fmt.Errorf("check existing release deployment: %w", err)
	}
	if existing > 0 {
		return nil
	}

	var install app.Install
	if err := a.db.WithContext(ctx).Where(app.Install{ID: flow.OwnerID, OrgID: flow.OrgID}).First(&install).Error; err != nil {
		return fmt.Errorf("load install: %w", err)
	}

	deployedBuilds, err := a.managedInstallBuilds(ctx, install.ID)
	if err != nil {
		return err
	}

	var release *app.AppRelease
	if releaseID := workflowReleaseID(flow); releaseID != "" {
		var selected app.AppRelease
		if err := a.db.WithContext(ctx).Where(app.AppRelease{
			ID: releaseID, OrgID: install.OrgID, AppID: install.AppID, AppConfigID: install.AppConfigID,
			Status: app.AppReleaseStatusReady,
		}).First(&selected).Error; err != nil {
			return fmt.Errorf("load selected release: %w", err)
		}
		if !sameBuildSet(selected.ComponentBuildIDs, deployedBuilds) {
			return fmt.Errorf("selected release component builds do not match active install deploys")
		}
		release = &selected
	} else {
		sandboxBuildID, err := a.managedSandboxBuildID(ctx, install)
		if err != nil {
			return err
		}
		if sandboxBuildID == "" {
			return nil
		}
		var releases []app.AppRelease
		if err := a.db.WithContext(ctx).Where(app.AppRelease{
			OrgID: install.OrgID, AppID: install.AppID, AppConfigID: install.AppConfigID,
			Status: app.AppReleaseStatusReady,
		}).Order("created_at DESC").Find(&releases).Error; err != nil {
			return fmt.Errorf("load candidate releases: %w", err)
		}
		for idx := range releases {
			if releases[idx].SandboxBuildID == sandboxBuildID && sameBuildSet(releases[idx].ComponentBuildIDs, deployedBuilds) {
				release = &releases[idx]
				break
			}
		}
	}
	if release == nil {
		return nil
	}
	sandboxMatches, err := a.managedSandboxMatches(ctx, install, *release)
	if err != nil {
		return err
	}
	if !sandboxMatches {
		return fmt.Errorf("selected release sandbox does not match the active install sandbox")
	}

	finishedAt := flow.FinishedAt.UTC()
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	}
	startedAt := flow.StartedAt.UTC()
	if startedAt.IsZero() {
		startedAt = finishedAt
	}

	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		operatingModel, err := managedInstallOperatingModel(tx, install, flow)
		if err != nil {
			return err
		}
		var previous app.InstallReleaseDeployment
		err = tx.Where(app.InstallReleaseDeployment{
			OrgID: install.OrgID, InstallID: install.ID, Status: app.InstallDeploymentStatusSucceeded,
		}).Order("finished_at DESC, created_at DESC").First(&previous).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load previous release deployment: %w", err)
		}
		deployment := app.InstallReleaseDeployment{
			OrgID: install.OrgID, InstallID: install.ID, ReleaseID: release.ID,
			OperatingModelID: operatingModel.ID, Method: app.InstallDeploymentMethodVendorManaged,
			Actor: "vendor", Executor: "temporal", OperationID: flow.ID,
			ResultDirective: "applied", Status: app.InstallDeploymentStatusSucceeded,
			StartedAt: startedAt, FinishedAt: &finishedAt,
		}
		var update app.InstallAppConfigVersion
		updateID := workflowMetadataValue(flow, "install_config_update_id")
		if err := tx.Where(app.InstallAppConfigVersion{ID: updateID, InstallID: install.ID}).First(&update).Error; err == nil {
			deployment.InstallAppConfigVersionID = &update.ID
			if update.OperatingModelID != nil {
				deployment.OperatingModelID = *update.OperatingModelID
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load release update: %w", err)
		}
		if previous.ID != "" {
			deployment.PreviousReleaseID = previous.ReleaseID
		}
		if err := tx.Create(&deployment).Error; err != nil {
			return fmt.Errorf("record managed release deployment: %w", err)
		}
		return nil
	})
}

func (a *Activities) managedInstallBuilds(ctx context.Context, installID string) (map[string]string, error) {
	var components []app.InstallComponent
	if err := a.db.WithContext(ctx).Where(app.InstallComponent{InstallID: installID}).Find(&components).Error; err != nil {
		return nil, fmt.Errorf("load install components: %w", err)
	}
	builds := make(map[string]string, len(components))
	for _, component := range components {
		var deploy app.InstallDeploy
		err := a.db.WithContext(ctx).Preload("ComponentBuild").Where(app.InstallDeploy{
			InstallComponentID: component.ID, Status: app.InstallDeployStatusActive,
		}).Order("applied_at DESC, created_at DESC").First(&deploy).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("load active deploy for component %s: %w", component.ID, err)
		}
		builds[deploy.ComponentBuild.ComponentConfigConnectionID] = deploy.ComponentBuildID
	}
	return builds, nil
}

func (a *Activities) managedSandboxBuildID(ctx context.Context, install app.Install) (string, error) {
	var runs []app.InstallSandboxRun
	if err := a.db.WithContext(ctx).Where(app.InstallSandboxRun{
		OrgID: install.OrgID, InstallID: install.ID, Status: app.SandboxRunStatusActive,
	}).Order("applied_at DESC, created_at DESC").Find(&runs).Error; err != nil {
		return "", fmt.Errorf("load install sandbox runs: %w", err)
	}
	for _, run := range runs {
		if run.AppSandboxBuildID != nil && *run.AppSandboxBuildID != "" {
			return *run.AppSandboxBuildID, nil
		}
	}
	return "", nil
}

func (a *Activities) managedSandboxMatches(ctx context.Context, install app.Install, release app.AppRelease) (bool, error) {
	var run app.InstallSandboxRun
	if err := a.db.WithContext(ctx).Where(app.InstallSandboxRun{
		OrgID: install.OrgID, InstallID: install.ID, Status: app.SandboxRunStatusActive,
	}).Order("applied_at DESC, created_at DESC").First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("load active install sandbox run: %w", err)
	}
	if run.AppSandboxBuildID != nil && *run.AppSandboxBuildID != "" {
		return *run.AppSandboxBuildID == release.SandboxBuildID, nil
	}
	var member app.AppReleaseMember
	if err := a.db.WithContext(ctx).Where(app.AppReleaseMember{
		OrgID: install.OrgID, ReleaseID: release.ID, Kind: "sandbox", BuildID: release.SandboxBuildID,
	}).First(&member).Error; err != nil {
		return false, fmt.Errorf("load selected release sandbox member: %w", err)
	}
	if err := a.db.WithContext(ctx).Where(app.AppSandboxConfig{
		ID: run.AppSandboxConfigID, OrgID: install.OrgID, AppID: install.AppID,
	}).First(&app.AppSandboxConfig{}).Error; err != nil {
		return false, fmt.Errorf("load active install sandbox config: %w", err)
	}
	var matchingMembers int64
	if err := a.db.WithContext(ctx).Model(&app.AppReleaseMember{}).Where(app.AppReleaseMember{
		OrgID: install.OrgID, Kind: "sandbox", AppSandboxConfigID: run.AppSandboxConfigID, ConfigDigest: member.ConfigDigest,
	}).Count(&matchingMembers).Error; err != nil {
		return false, fmt.Errorf("match active install sandbox config to release: %w", err)
	}
	return matchingMembers > 0, nil
}

func workflowReleaseID(flow app.Workflow) string {
	return workflowMetadataValue(flow, "app_release_id")
}

func workflowMetadataValue(flow app.Workflow, key string) string {
	value, ok := flow.Metadata[key]
	if !ok || value == nil {
		return ""
	}
	return *value
}

func recordsReleaseDeployment(workflowType app.WorkflowType) bool {
	switch workflowType {
	case app.WorkflowTypeProvision, app.WorkflowTypeReprovision, app.WorkflowTypeAppBranchConfigUpdate,
		app.WorkflowTypeDeployComponents, app.WorkflowTypeManualDeploy,
		app.WorkflowTypeComponentEnabled, app.WorkflowTypeComponentDisabled:
		return true
	default:
		return false
	}
}

func sameBuildSet(expected, actual map[string]string) bool {
	if len(expected) != len(actual) {
		return false
	}
	for key, value := range expected {
		if actual[key] != value {
			return false
		}
	}
	return true
}

func managedInstallOperatingModel(tx *gorm.DB, install app.Install, flow app.Workflow) (*app.InstallOperatingModel, error) {
	var operatingModel app.InstallOperatingModel
	err := tx.Where(app.InstallOperatingModel{OrgID: install.OrgID, InstallID: install.ID}).First(&operatingModel).Error
	if err == nil {
		return &operatingModel, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("load install operating model: %w", err)
	}
	operatingModel = app.InstallOperatingModel{
		CreatedByID: flow.CreatedByID, OrgID: install.OrgID, InstallID: install.ID,
		Connectivity: app.InstallConnectivityConnected, ReleaseSelection: app.InstallReleaseSelectionVendor,
		ApprovalAuthority: app.InstallAuthorityVendor,
		Telemetry:         app.InstallTelemetryLive,
	}
	if err := tx.Create(&operatingModel).Error; err != nil {
		return nil, fmt.Errorf("create managed install operating model: %w", err)
	}
	return &operatingModel, nil
}
