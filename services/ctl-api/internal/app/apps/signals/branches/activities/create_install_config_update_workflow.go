package activities

import (
	"context"
	"fmt"

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
	update, wf, err := a.installHelpers.StartInstallAppConfigUpdate(ctx, installshelpers.StartInstallAppConfigUpdateInput{
		InstallID: input.InstallID, NewAppConfigID: input.NewAppConfigID,
		AppBranchRunID: input.AppBranchRunID, InstallGroupID: input.InstallGroupID,
		PlanOnly: input.PlanOnly, Callback: input.Callback,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create install config update workflow: %w", err)
	}

	return &CreateInstallAppConfigVersionWorkflowOutput{
		WorkflowID:                wf.ID,
		InstallAppConfigVersionID: update.ID,
	}, nil
}
