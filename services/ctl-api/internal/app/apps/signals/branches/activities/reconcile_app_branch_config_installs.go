package activities

import (
	"context"
)

type ReconcileAppBranchConfigInstallsInput struct {
	AppBranchID       string `json:"app_branch_id" validate:"required"`
	AppBranchConfigID string `json:"app_branch_config_id" validate:"required"`
}

type ReconciledAppBranchInstall struct {
	InstallID      string `json:"install_id"`
	InstallGroupID string `json:"install_group_id"`

	ShouldSignal bool `json:"should_signal"`
}

type ReconcileAppBranchConfigInstallsOutput struct {
	Installs []ReconciledAppBranchInstall `json:"installs"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) ReconcileAppBranchConfigInstalls(ctx context.Context, input *ReconcileAppBranchConfigInstallsInput) (*ReconcileAppBranchConfigInstallsOutput, error) {
	return &ReconcileAppBranchConfigInstallsOutput{}, nil
}
