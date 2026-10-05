package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
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
	deferDiff, err := a.applyInstallUpdatePolicy(ctx, input)
	if err != nil {
		return nil, err
	}
	update, err := a.installHelpers.CreateAppBranchConfigUpdateWorkflow(ctx, installhelpers.AppBranchConfigUpdateInput{
		InstallID:      input.InstallID,
		NewAppConfigID: input.NewAppConfigID,
		AppBranchRunID: input.AppBranchRunID,
		InstallGroupID: input.InstallGroupID,
		PlanOnly:       input.PlanOnly,
		Callback:       input.Callback,
		DeferDiff:      deferDiff,
	})
	if err != nil {
		return nil, err
	}

	return &CreateInstallAppConfigVersionWorkflowOutput{
		WorkflowID:                update.WorkflowID,
		InstallAppConfigVersionID: update.InstallAppConfigVersionID,
	}, nil
}

func (a *Activities) applyInstallUpdatePolicy(ctx context.Context, input *CreateInstallAppConfigVersionWorkflowInput) (bool, error) {
	policy, err := a.installHelpers.InstallUpdatePolicyForRun(ctx, input.AppBranchRunID)
	if err != nil {
		return false, err
	}
	prior, err := a.installHelpers.PriorInFlightInstallUpdate(ctx, input.InstallID, "")
	if err != nil {
		return false, err
	}
	if prior == nil {
		return false, nil
	}
	if policy == app.InstallUpdatePolicyQueue {
		return true, nil
	}
	if err := a.CancelInstallWorkflow(ctx, &CancelInstallWorkflowInput{WorkflowID: prior.WorkflowID}); err != nil {
		return false, err
	}
	if err := a.installHelpers.MarkInstallSupersededForInstall(ctx, prior.AppBranchRunID, input.InstallID, input.AppBranchRunID); err != nil {
		return false, err
	}
	return false, nil
}
