package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallConfigUpdateWorkflowInput struct {
	InstallID      string `json:"install_id"`
	NewAppConfigID string `json:"new_app_config_id"`
	AppBranchRunID string `json:"app_branch_run_id"`
	InstallGroupID string `json:"install_group_id"`
	PlanOnly       bool   `json:"plan_only"`
}

type CreateInstallConfigUpdateWorkflowOutput struct {
	WorkflowID            string `json:"workflow_id"`
	InstallConfigUpdateID string `json:"install_config_update_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallConfigUpdateWorkflow(ctx context.Context, input *CreateInstallConfigUpdateWorkflowInput) (*CreateInstallConfigUpdateWorkflowOutput, error) {
	// Get the install to find its current AppConfigID
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	// Create the install workflow via install helpers (handles approval config, metadata, etc.)
	metadata := map[string]string{
		"new_app_config_id": input.NewAppConfigID,
		"app_branch_run_id": input.AppBranchRunID,
		"install_group_id":  input.InstallGroupID,
	}

	workflow, err := a.installHelpers.CreateWorkflow(
		ctx,
		input.InstallID,
		app.WorkflowTypeAppBranchConfigUpdate,
		metadata,
		input.PlanOnly,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install config update workflow: %w", err)
	}

	// Create the InstallConfigUpdate tracking record
	update := app.InstallConfigUpdate{
		AppBranchRunID: input.AppBranchRunID,
		InstallGroupID: input.InstallGroupID,
		InstallID:      input.InstallID,
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		WorkflowID:     &workflow.ID,
		Status:         "pending",
	}
	if err := a.db.WithContext(ctx).Create(&update).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config update: %w", err)
	}

	return &CreateInstallConfigUpdateWorkflowOutput{
		WorkflowID:            workflow.ID,
		InstallConfigUpdateID: update.ID,
	}, nil
}
