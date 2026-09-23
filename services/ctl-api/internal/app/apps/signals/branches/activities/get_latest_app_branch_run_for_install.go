package activities

import (
	"context"

	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

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
	latest, err := a.installHelpers.LatestAppBranchRunForInstall(ctx, input.AppBranchID, input.InstallID)
	if err != nil {
		return nil, installhelpers.AppBranchRunResolveActivityError(err)
	}

	return &GetLatestAppBranchRunForInstallOutput{
		AppBranchRunID: latest.AppBranchRunID,
		AppConfigID:    latest.AppConfigID,
		InstallGroupID: latest.InstallGroupID,
		AlreadyCurrent: latest.AlreadyCurrent,
	}, nil
}
