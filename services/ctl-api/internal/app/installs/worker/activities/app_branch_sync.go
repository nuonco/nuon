package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
)

// The app-branch-changed signal runs on the install-signals queue, so these
// activities have to be registered on the installs worker. They mirror the
// app-branch signal activities in apps/signals/branches/activities; both sets
// delegate to the same install helpers.

type EnsureInstallAppBranchInput struct {
	InstallID   string `json:"install_id" validate:"required"`
	AppBranchID string `json:"app_branch_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) EnsureInstallAppBranch(ctx context.Context, input *EnsureInstallAppBranchInput) error {
	return a.helpers.EnsureInstallAppBranch(ctx, input.InstallID, input.AppBranchID)
}

type GetLatestAppBranchRunForInstallInput struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
	InstallID   string `json:"install_id" validate:"required"`
}

type GetLatestAppBranchRunForInstallOutput struct {
	AppBranchRunID string `json:"app_branch_run_id,omitempty"`
	AppConfigID    string `json:"app_config_id,omitempty"`
	InstallGroupID string `json:"install_group_id,omitempty"`
	AlreadyCurrent bool   `json:"already_current"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetLatestAppBranchRunForInstall(ctx context.Context, input *GetLatestAppBranchRunForInstallInput) (*GetLatestAppBranchRunForInstallOutput, error) {
	latest, err := a.helpers.LatestAppBranchRunForInstall(ctx, input.AppBranchID, input.InstallID)
	if err != nil {
		return nil, err
	}

	return &GetLatestAppBranchRunForInstallOutput{
		AppBranchRunID: latest.AppBranchRunID,
		AppConfigID:    latest.AppConfigID,
		InstallGroupID: latest.InstallGroupID,
		AlreadyCurrent: latest.AlreadyCurrent,
	}, nil
}

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
	update, err := a.helpers.CreateAppBranchConfigUpdateWorkflow(ctx, helpers.AppBranchConfigUpdateInput{
		InstallID:      input.InstallID,
		NewAppConfigID: input.NewAppConfigID,
		AppBranchRunID: input.AppBranchRunID,
		InstallGroupID: input.InstallGroupID,
		PlanOnly:       input.PlanOnly,
		Callback:       input.Callback,
	})
	if err != nil {
		return nil, err
	}

	return &CreateInstallAppConfigVersionWorkflowOutput{
		WorkflowID:                update.WorkflowID,
		InstallAppConfigVersionID: update.InstallAppConfigVersionID,
	}, nil
}
