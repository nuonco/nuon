package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanagedservice "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/service"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
)

type CreateInstallAppConfigVersionWorkflowInput struct {
	InstallID      string       `json:"install_id"`
	NewAppConfigID string       `json:"new_app_config_id"`
	AppBranchRunID string       `json:"app_branch_run_id"`
	InstallGroupID string       `json:"install_group_id"`
	PlanOnly       bool         `json:"plan_only"`
	Callback       callback.Ref `json:"callback,omitempty"`
}

type CreateInstallAppConfigVersionWorkflowOutput struct {
	WorkflowID                string `json:"workflow_id"`
	InstallAppConfigVersionID string `json:"install_config_update_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallAppConfigVersionWorkflow(ctx context.Context, input *CreateInstallAppConfigVersionWorkflowInput) (*CreateInstallAppConfigVersionWorkflowOutput, error) {
	startInput := installshelpers.StartInstallAppConfigUpdateInput{
		InstallID: input.InstallID, NewAppConfigID: input.NewAppConfigID,
		AppBranchRunID: input.AppBranchRunID, InstallGroupID: input.InstallGroupID,
		PlanOnly: input.PlanOnly, Callback: input.Callback,
	}

	var install app.Install
	if err := a.db.WithContext(ctx).
		Preload("OperatingModel").
		Where(app.Install{ID: input.InstallID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to load install operating model: %w", err)
	}
	if !install.AppBranchUpdateEligible() {
		return nil, fmt.Errorf("install %s operating model does not allow app branch updates", install.ID)
	}
	if install.OperatingModel != nil && install.OperatingModel.ApprovalAuthority == app.InstallAuthorityCustomer {
		release, err := customermanagedservice.CreateAppReleaseForConfig(
			ctx,
			a.db,
			a.helpers,
			a.blobSvc,
			install.OrgID,
			install.AppID,
			input.NewAppConfigID,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create app release for branch run: %w", err)
		}
		startInput.AppReleaseID = release.ID
		startInput.OperatingModelID = install.OperatingModel.ID
		startInput.ReleaseComponentBuildIDs = release.ComponentBuildIDs
		startInput.ReleaseSandboxBuildID = release.SandboxBuildID
	}

	update, wf, err := a.installHelpers.StartInstallAppConfigUpdate(ctx, startInput)
	if err != nil {
		return nil, fmt.Errorf("unable to create install config update workflow: %w", err)
	}

	return &CreateInstallAppConfigVersionWorkflowOutput{
		WorkflowID:                wf.ID,
		InstallAppConfigVersionID: update.ID,
	}, nil
}
